package common

import (
	"html/template"
	"net/http"
	"sync"
	"time"
)

// TODO: Allow resources in spots other than /static/ and possibly even external domains (e.g. CDNs)
// TODO: Preload Trumboyg on Cosora on the forum list
type Header struct {
	Title      string
	NoticeList []string
	Scripts    []string
	//PreloadScripts []string
	Stylesheets []string
	Widgets     PageWidgets
	Site        *site
	Settings    SettingMap
	Themes      map[string]*Theme // TODO: Use a slice containing every theme instead of the main map for speed?
	Theme       *Theme
	//TemplateName string // TODO: Use this to move template calls to the router rather than duplicating them over and over and over?
	// TODO: Use a pointer here
	CurrentUser User // TODO: Deprecate CurrentUser on the page structs
	Zone        string
	MetaDesc    string
	Writer      http.ResponseWriter
	ExtData     ExtData
}

func (header *Header) AddScript(name string) {
	header.Scripts = append(header.Scripts, name)
}

/*func (header *Header) PreloadScript(name string) {
	header.PreloadScripts = append(header.PreloadScripts, name)
}*/

func (header *Header) AddSheet(name string) {
	header.Stylesheets = append(header.Stylesheets, name)
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
	Header      *Header
	ItemList    []interface{}
	Something   interface{}
}

type Paginator struct {
	PageList []int
	Page     int
	LastPage int
}

type TopicPage struct {
	*Header
	ItemList []ReplyUser
	Topic    TopicUser
	Poll     Poll
	Page     int
	LastPage int
}

type TopicListPage struct {
	*Header
	TopicList    []*TopicsRow
	ForumList    []Forum
	DefaultForum int
	Paginator
}

type ForumPage struct {
	*Header
	ItemList []*TopicsRow
	Forum    *Forum
	Paginator
}

type ForumsPage struct {
	*Header
	ItemList []Forum
}

type ProfilePage struct {
	*Header
	ItemList     []ReplyUser
	ProfileOwner User
}

type CreateTopicPage struct {
	*Header
	ItemList []Forum
	FID      int
}

type IPSearchPage struct {
	*Header
	ItemList map[int]*User
	IP       string
}

type EmailListPage struct {
	*Header
	ItemList  []Email
	Something interface{}
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
	Header      *Header
	Stats       PanelStats
	Zone        string
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
	*Header
	Stats     PanelStats
	Zone      string
	GridItems []GridElement
}

type PanelSetting struct {
	*Setting
	FriendlyName string
}

type PanelSettingPage struct {
	*Header
	Stats    PanelStats
	Zone     string
	ItemList []OptionLabel
	Setting  *PanelSetting
}

type PanelTimeGraph struct {
	Series []int64 // The counts on the left
	Labels []int64 // unixtimes for the bottom, gets converted into 1:00, 2:00, etc. with JS
}

type PanelAnalyticsItem struct {
	Time  int64
	Count int64
}

type PanelAnalyticsPage struct {
	Title        string
	CurrentUser  User
	Header       *Header
	Stats        PanelStats
	Zone         string
	PrimaryGraph PanelTimeGraph
	ViewItems    []PanelAnalyticsItem
	TimeRange    string
}

type PanelAnalyticsRoutesItem struct {
	Route string
	Count int
}

type PanelAnalyticsRoutesPage struct {
	Title       string
	CurrentUser User
	Header      *Header
	Stats       PanelStats
	Zone        string
	ItemList    []PanelAnalyticsRoutesItem
	TimeRange   string
}

type PanelAnalyticsAgentsItem struct {
	Agent         string
	FriendlyAgent string
	Count         int
}

type PanelAnalyticsAgentsPage struct {
	Title       string
	CurrentUser User
	Header      *Header
	Stats       PanelStats
	Zone        string
	ItemList    []PanelAnalyticsAgentsItem
	TimeRange   string
}

type PanelAnalyticsRoutePage struct {
	Title        string
	CurrentUser  User
	Header       *Header
	Stats        PanelStats
	Zone         string
	Route        string
	PrimaryGraph PanelTimeGraph
	ViewItems    []PanelAnalyticsItem
	TimeRange    string
}

type PanelAnalyticsAgentPage struct {
	Title         string
	CurrentUser   User
	Header        *Header
	Stats         PanelStats
	Zone          string
	Agent         string
	FriendlyAgent string
	PrimaryGraph  PanelTimeGraph
	TimeRange     string
}

type PanelThemesPage struct {
	*Header
	Stats         PanelStats
	Zone          string
	PrimaryThemes []*Theme
	VariantThemes []*Theme
}

type PanelMenuListItem struct {
	Name      string
	ID        int
	ItemCount int
}

type PanelMenuListPage struct {
	*Header
	Stats    PanelStats
	Zone     string
	ItemList []PanelMenuListItem
}

type PanelMenuPage struct {
	*Header
	Stats    PanelStats
	Zone     string
	MenuID   int
	ItemList []MenuItem
}

type PanelMenuItemPage struct {
	*Header
	Stats PanelStats
	Zone  string
	Item  MenuItem
}

type PanelUserPage struct {
	*Header
	Stats    PanelStats
	Zone     string
	ItemList []*User
	Paginator
}

type PanelGroupPage struct {
	Title       string
	CurrentUser User
	Header      *Header
	Stats       PanelStats
	Zone        string
	ItemList    []GroupAdmin
	Paginator
}

type PanelEditGroupPage struct {
	Title       string
	CurrentUser User
	Header      *Header
	Stats       PanelStats
	Zone        string
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
	Header      *Header
	Stats       PanelStats
	Zone        string
	ID          int
	Name        string
	Desc        string
	Active      bool
	Preset      string
	Groups      []GroupForumPermPreset
}

type NameLangToggle struct {
	Name    string
	LangStr string
	Toggle  bool
}

type PanelEditForumGroupPage struct {
	Title       string
	CurrentUser User
	Header      *Header
	Stats       PanelStats
	Zone        string
	ForumID     int
	GroupID     int
	Name        string
	Desc        string
	Active      bool
	Preset      string
	Perms       []NameLangToggle
}

type PanelEditGroupPermsPage struct {
	Title       string
	CurrentUser User
	Header      *Header
	Stats       PanelStats
	Zone        string
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
	Header      *Header
	Stats       PanelStats
	Zone        string
	Backups     []BackupItem
}

type PageLogItem struct {
	Action    template.HTML
	IPAddress string
	DoneAt    string
}

type PanelLogsPage struct {
	Title       string
	CurrentUser User
	Header      *Header
	Stats       PanelStats
	Zone        string
	Logs        []PageLogItem
	Paginator
}

type PageRegLogItem struct {
	RegLogItem
	ParsedReason string
}

type PanelRegLogsPage struct {
	Title       string
	CurrentUser User
	Header      *Header
	Stats       PanelStats
	Zone        string
	Logs        []PageRegLogItem
	Paginator
}

type PanelDebugPage struct {
	Title       string
	CurrentUser User
	Header      *Header
	Stats       PanelStats
	Zone        string
	GoVersion   string
	DBVersion   string
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

// TODO: Write a test for this
func DefaultHeader(w http.ResponseWriter) *Header {
	return &Header{Site: Site, Theme: Themes[fallbackTheme], CurrentUser: GuestUser, Writer: w}
}
