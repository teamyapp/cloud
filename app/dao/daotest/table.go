package daotest

const AllocatedRangeTableName = "allocatedRange"
const ChunkMetadataTableName = "chunkMetadata"
const FileMetadataTableName = "fileMetadata"
const OperationTableName = "operation"
const OperationRelationTableName = "operationRelation"
const PermissionTableName = "permission"
const ResourceTableName = "resource"
const ResourceRelationTableName = "resourceRelation"
const ResourceTypeTableName = "resourceType"
const ServiceAccountTableName = "serviceAccount"
const SignInSessionTableName = "signInSession"
const UploadSessionTableName = "uploadSession"
const UserGroupTableName = "userGroup"
const UserGroupMemberTableName = "userGroupMember"
const UserLinkTableName = "userLink"

type Table struct {
	rows []interface{}
}

func newTable() *Table {
	return &Table{rows: []interface{}{}}
}
