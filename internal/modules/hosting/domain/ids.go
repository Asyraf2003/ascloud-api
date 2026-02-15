package domain

type SiteID string
type UploadID string
type ReleaseID string

func (id SiteID) String() string    { return string(id) }
func (id UploadID) String() string  { return string(id) }
func (id ReleaseID) String() string { return string(id) }
