package domain

type Link struct {
	URL   string
	Title string
}

type SettingAggregate struct {
	Title      string
	Navigation []Link
}
