package storage

import (
	"fmt"
	"io"

	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

func NewGremlinStorage(connectionString string) (GitStorage, error) {
	conn, err := gremlin.NewDriverRemoteConnection(connectionString)
	if err != nil {
		return nil, err
	}
	return &GremlinStorage{
		conn: conn,
	}, nil
}

type GremlinStorage struct {
	conn *gremlin.DriverRemoteConnection // Object storage

	// Everything else (nil for now, since we shouldn't be using any of them)
	storer.ReferenceStorer
}

func (s *GremlinStorage) Close() error {
	s.conn.Close()
	return nil
}

// Object properties:
// Hash (OID)
// Size
// Type
// Commit stuff (author, committer, signature, mergetag, message)
// Tag stuff
// Blob (content)
// Trees only have edge data

// Edge types:
// Parent (property: order)
// Tree
// Entry (properties: name, mode)

func (s *GremlinStorage) NewEncodedObject() plumbing.EncodedObject {
	return &plumbing.MemoryObject{}
}

func (s *GremlinStorage) SetEncodedObject(obj plumbing.EncodedObject) (plumbing.Hash, error) {
	oid := obj.Hash()
	oidStr := oid.String()
	objType := obj.Type().String()
	g := gremlin.Traversal_().WithRemote(s.conn)

	query := g.AddV(objType).Property("oid", oidStr)

	switch obj.Type() {
	case plumbing.CommitObject:
		commit, err := object.DecodeCommit(s, obj)
		if err != nil {
			return plumbing.ZeroHash, err
		}

		query.Property("author", commit.Author,
			"committer", commit.Committer,
			"gpgsig", commit.PGPSignature,
			"mergetag", commit.MergeTag,
			"message", commit.Message)
		v, err := query.Next()
		if err != nil {
			return plumbing.ZeroHash, err
		}

		// Add parents
		for i, parent := range commit.ParentHashes {
			_, err = g.
				V(v).As("source").
				V("commit").Has("oid", parent.String()).As("target").
				AddE("parent", "order", i).From("source").To("target").Next()
			if err != nil {
				return plumbing.ZeroHash, err
			}
		}

		// Add tree
		_, err = g.
			V(v).As("source").
			V("tree").Has("oid", commit.TreeHash.String()).As("target").
			AddE("tree").From("source").To("target").Next()
		if err != nil {
			return plumbing.ZeroHash, err
		}
	case plumbing.TagObject:
		return plumbing.ZeroHash, fmt.Errorf("not implemented")
	case plumbing.TreeObject:
		tree, err := object.DecodeTree(s, obj)
		if err != nil {
			return plumbing.ZeroHash, err
		}

		// Write the tree
		v, err := query.Next()
		if err != nil {
			return plumbing.ZeroHash, err
		}

		// Connect to tree entries
		for _, ent := range tree.Entries {
			_, err = g.
				V(v).As("root").
				V().Has("oid", ent.Hash.String()).As("entry").
				AddE(ent.Name, "mode", ent.Mode.String()).Next()
			if err != nil {
				return plumbing.ZeroHash, err
			}
		}
	case plumbing.BlobObject:
		reader, err := obj.Reader()
		if err != nil {
			return plumbing.ZeroHash, err
		}

		buf, err := io.ReadAll(reader)
		if err != nil {
			return plumbing.ZeroHash, err
		} else if len(buf) < int(obj.Size()) {
			return plumbing.ZeroHash, fmt.Errorf("incorrect number of bytes in object (expected %d, got %d)", obj.Size(), len(buf))
		}

		_, err = query.Property("content", string(buf)).Next()
		if err != nil {
			return plumbing.ZeroHash, err
		}
	default:
		return plumbing.ZeroHash, fmt.Errorf("invalid object type %s", objType)
	}

	return oid, nil
}

func (s *GremlinStorage) EncodedObject(objType plumbing.ObjectType, oid plumbing.Hash) (plumbing.EncodedObject, error) {
	oidStr := oid.String()
	g := gremlin.Traversal_().WithRemote(s.conn)

	props := []interface{}{}
	if objType > plumbing.InvalidObject {
		props = append(props, objType.String())
	}

	props = append(props, "oid", oidStr)

	res, err := g.V().Has(props...).Next()
	if err != nil {
		return nil, plumbing.ErrObjectNotFound
	}

	v, err := res.GetVertex()
	if err != nil {
		return nil, err
	}

	obj := s.NewEncodedObject()
	outType, err := plumbing.ParseObjectType(v.Label)
	if err != nil {
		return nil, err
	}
	obj.SetType(outType)

	switch outType {
	case plumbing.CommitObject:
		return nil, fmt.Errorf("not implemented")
	case plumbing.TagObject:
		return nil, fmt.Errorf("not implemented")
	case plumbing.TreeObject:
		// Get all the edges and vertices connected to them
		// t := &object.Tree{}

		query := g.V(v.Id).OutE().As("entry").OtherV().As("object").Select("entry", "object")
		var keepGoing bool
		for keepGoing, err = query.HasNext(); keepGoing; {
			res, err := query.Next()
			if err != nil {
				return nil, err
			}

			resMap, ok := res.Data.(map[interface{}]interface{})
			if !ok {
				return nil, fmt.Errorf("could not get tree entry")
			}
			_, ok = resMap["entry"].(*gremlin.Edge)
			if !ok {
				return nil, fmt.Errorf("could not get tree entry edge")
			}

			_, ok = resMap["object"].(*gremlin.Edge)
			if !ok {
				return nil, fmt.Errorf("could not get tree entry object")
			}
		}

		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("not implemented")
	case plumbing.BlobObject:
		// Get content
		res, err = g.V(v.Id).Values("content").Next()
		if err != nil {
			return nil, err
		}

		content := res.GetString()

		obj.SetSize(int64(len(content)))

		objWriter, err := obj.Writer()
		if err != nil {
			return nil, err
		}
		objWriter.Write([]byte(content))

		return obj, nil
	default:
		return nil, fmt.Errorf("invalid object type %s", v.Label)
	}
}

func (s *GremlinStorage) IterEncodedObjects(objType plumbing.ObjectType) (storer.EncodedObjectIter, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *GremlinStorage) HasEncodedObject(oid plumbing.Hash) error {
	oidStr := oid.String()
	g := gremlin.Traversal_().WithRemote(s.conn)

	res, err := g.V().HasLabel("oid", oidStr).Next()
	if res != nil && err != nil {
		return nil
	} else {
		return plumbing.ErrObjectNotFound
	}
}

func (s *GremlinStorage) EncodedObjectSize(oid plumbing.Hash) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (s *GremlinStorage) AddAlternate(remote string) error {
	// No alternates support
	return fmt.Errorf("alternates are not supported")
}
