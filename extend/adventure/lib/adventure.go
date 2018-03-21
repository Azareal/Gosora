package adventure

// We're experimenting with struct tags here atm
type Adventure struct {
	ID        int    `schema:"name=aid;primary;auto"`
	Name      string `schema:"name=name;type=short_text"`
	Desc      string `schema:"name=desc;type=text"`
	CreatedBy int    `schema:"name=createdBy"`
	//CreatedBy int `schema:"name=createdBy;relatesTo=users.uid"`
}

// TODO: Should we add a table interface?
func (adventure *Adventure) GetTable() string {
	return "adventure"
}
