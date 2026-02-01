package entity

import "rakit-tiket-be/pkg/util"

type (
	UUID string
)

func MakeUUID(strs ...string) UUID {
	return UUID(util.MakeUUID(strs...))
}

func (e UUID) String() string {
	return string(e)
}

type (
	UUIDs    []UUID
	UUIDMap  map[string]UUID
	UUIDsMap map[string]UUIDs
)

func (e UUIDs) Strings() []string {
	var uuids []string
	for _, uuid := range e {
		uuids = append(uuids, uuid.String())
	}
	return uuids
}

func (e UUID) IsValidUUID() bool {
	return util.IsValidUUID(e.String())
}
