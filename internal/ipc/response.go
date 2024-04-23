package ipc

type GetOidResponse struct {
	Key
	Oid          ObjectId
	DeltaBaseOid ObjectId
	DiskSize     int64
	Size         uint32
	Whence       uint16
	Type         ObjectType
}
