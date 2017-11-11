package common

import (
	"html/template"
	"sync"
	"time"
)

type HeaderVars struct {
	NoticeList  []string
	Scripts     []string
	Stylesheets []string
	Widgets     PageWidgets
	Site        *site
	Settings    map[string]interface{}
	Themes      map[string]Theme // TODO: Use a slice containing every theme instead of the main map for speed
	ThemeName   string
	//TemplateName string // TODO: Use this to move template calls to the router rather than duplicating them over and over and over?
	ExtData ExtData
}

// TODO: Add this to routes which don't use templates. E.g. Json APIs.
type HeaderLite struct {
	Site     *site
	Settings SettingMap
	ExtData  ExtData
}

type PageWidgets struct {
	LeftSidebar  template.HTML
	RightSidebar template.HTML
}

// TODO: Add a ExtDataHolder interface with methods for manipulating the contents?
// ? - Could we use a sync.Map instead?
type ExtData struct {
	Items map[string]interface{} // Key: pluginname
	sync.RWMutex
}

type Page struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	ItemList    []interface{}
	Something   interface{}
}

type TopicPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	ItemList    []ReplyUser
	Topic       TopicUser
	Page        int
	LastPage    int
}

type TopicsPage struct {
	Title        string
	CurrentUser  User
	Header       *HeaderVars
	TopicList    []*TopicsRow
	ForumList    []Forum
	DefaultForum int
}

type ForumPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	ItemList    []*TopicsRow
	Forum       *Forum
	Page        int
	LastPage    int
}

type ForumsPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	ItemList    []Forum
}

type ProfilePage struct {
	Title        string
	CurrentUser  User
	Header       *HeaderVars
	ItemList     []ReplyUser
	ProfileOwner User
}

type CreateTopicPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	ItemList    []Forum
	FID         int
}

type IPSearchPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	ItemList    map[int]*User
	IP          string
}

type PanelStats struct {
	Users       int
	Groups      int
	Forums      int
	Settings    int
	WordFilters int
	Themes      int
	Reports     int
}

type PanelPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	Stats       PanelStats
	ItemList    []interface{}
	Something   interface{}
}

type GridElement struct {
	ID         string
	Body       string
	Order      int // For future use
	Class      string
	Background string
	TextColour string
	Note       string
}

type PanelDashboardPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	Stats       PanelStats
	GridItems   []GridElement
}

type PanelThemesPage struct {
	Title         string
	CurrentUser   User
	Header        *HeaderVars
	Stats         PanelStats
	PrimaryThemes []Theme
	VariantThemes []Theme
}

type PanelUserPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	Stats       PanelStats
	ItemList    []User
	PageList    []int
	Page        int
	LastPage    int
}

type PanelGroupPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	Stats       PanelStats
	ItemList    []GroupAdmin
	PageList    []int
	Page        int
	LastPage    int
}

type PanelEditGroupPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	Stats       PanelStats
	ID          int
	Name        string
	Tag         string
	Rank        string
	DisableRank bool
}

type GroupForumPermPreset struct {
	Group  *Group
	Preset string
}

type PanelEditForumPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	Stats       PanelStats
	ID          int
	Name        string
	Desc        string
	Active      bool
	Preset      string
	Groups      []GroupForumPermPreset
}

/*type NameLangPair struct {
	Name    string
	LangStr string
}*/

type NameLangToggle struct {
	Name    string
	LangStr string
	Toggle  bool
}

type PanelEditGroupPermsPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	Stats       PanelStats
	ID          int
	Name        string
	LocalPerms  []NameLangToggle
	GlobalPerms []NameLangToggle
}

type BackupItem struct {
	SQLURL string

	// TODO: Add an easier to parse format here for Gosora to be able to more easily reimport portions of the dump and to strip unnecessary data (e.g. table defs and parsed post data)

	Timestamp time.Time
}

type PanelBackupPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	Stats       PanelStats
	Backups     []BackupItem
}

type LogItem struct {
	Action    template.HTML
	IPAddress string
	DoneAt    string
}

type PanelLogsPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	Stats       PanelStats
	Logs        []LogItem
	PageList    []int
	Page        int
	LastPage    int
}

type PanelDebugPage struct {
	Title       string
	CurrentUser User
	Header      *HeaderVars
	Stats       PanelStats
	Uptime      string
	OpenConns   int
	DBAdapter   string
}

type PageSimple struct {
	Title     string
	Something interface{}
}

type AreYouSure struct {
	URL     string
	Message string
}

// This is mostly for errors.go, please create *HeaderVars on the spot instead of relying on this or the atomic store underlying it, if possible
// TODO: Write a test for this
func DefaultHeaderVar() *HeaderVars {
	return &HeaderVars{Site: Site, ThemeName: fallbackTheme}
}
