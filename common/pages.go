package common

import (
	"html/template"
	"net/http"
	"runtime"
	"sync"
	"time"

	p "github.com/Azareal/Gosora/common/phrases"
)

/*type HResource struct {
	Name string
	Hash string
}*/

// TODO: Allow resources in spots other than /s/ and possibly even external domains (e.g. CDNs)
// TODO: Preload Trumboyg on Cosora on the forum list
type Header struct {
	Title string
	//Title      []byte // Experimenting with []byte for increased efficiency, let's avoid converting too many things to []byte, as it involves a lot of extra boilerplate
	NoticeList      []string
	Scripts         []string
	PreScriptsAsync []string
	ScriptsAsync    []string
	//Preload []string
	Stylesheets []string
	Widgets     PageWidgets
	Site        *site
	Settings    SettingMap
	Themes      map[string]*Theme // TODO: Use a slice containing every theme instead of the main map for speed?
	Theme       *Theme
	//TemplateName string // TODO: Use this to move template calls to the router rather than duplicating them over and over and over?
	CurrentUser User // TODO: Deprecate CurrentUser on the page structs and use a pointer here
	Hooks       *HookTable
	Zone        string
	ZoneID      int
	ZoneData    interface{}
	Path        string
	MetaDesc    string
	//OGImage string
	OGDesc         string
	GoogSiteVerify string
	IsoCode        string
	LooseCSP       bool
	StartedAt      time.Time
	Elapsed1       string
	Writer         http.ResponseWriter
	ExtData        ExtData
}

func (h *Header) AddScript(name string) {
	if name[0] == '/' && name[1] != '/' {
		// TODO: Use a secondary static file map to avoid this concatenation?
		file, ok := StaticFiles.Get("/s/" + name)
		if ok {
			name = file.OName
		}
	}
	//log.Print("name:", name)
	h.Scripts = append(h.Scripts, name)
}

func (h *Header) AddPreScriptAsync(name string) {
	if name[0] == '/' && name[1] != '/' {
		file, ok := StaticFiles.Get("/s/" + name)
		if ok {
			name = file.OName
		}
	}
	h.PreScriptsAsync = append(h.PreScriptsAsync, name)
}

func (h *Header) AddScriptAsync(name string) {
	if name[0] == '/' && name[1] != '/' {
		file, ok := StaticFiles.Get("/s/" + name)
		if ok {
			name = file.OName
		}
	}
	h.ScriptsAsync = append(h.ScriptsAsync, name)
}

/*func (h *Header) Preload(name string) {
	h.Preload = append(h.Preload, name)
}*/

func (h *Header) AddSheet(name string) {
	if name[0] == '/' && name[1] != '/' {
		file, ok := StaticFiles.Get("/s/" + name)
		if ok {
			name = file.OName
		}
	}
	h.Stylesheets = append(h.Stylesheets, name)
}

func (h *Header) AddNotice(name string) {
	h.NoticeList = append(h.NoticeList, p.GetNoticePhrase(name))
}

// TODO: Add this to routes which don't use templates. E.g. Json APIs.
type HeaderLite struct {
	Site     *site
	Settings SettingMap
	Hooks    *HookTable
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
	*Header
	ItemList  []interface{}
	Something interface{}
}

type SimplePage struct {
	*Header
}

type ErrorPage struct {
	*Header
	Message string
}

type Paginator struct {
	PageList []int
	Page     int
	LastPage int
}

type CustomPagePage struct {
	*Header
	Page *CustomPage
}

type TopicCEditPost struct {
	ID     int
	Source string
	Ref    string
}
type TopicCAttachItem struct {
	ID       int
	ImgSrc   string
	Path     string
	FullPath string
}
type TopicCPollInput struct {
	Index int
	Place string
}

type TopicPage struct {
	*Header
	ItemList []*ReplyUser
	Topic    TopicUser
	Forum    *Forum
	Poll     Poll
	Paginator
}

type TopicListSort struct {
	SortBy    string // lastupdate, mostviewed, mostviewedtoday, mostviewedthisweek, mostviewedthismonth
	Ascending bool
}

type TopicListPage struct {
	*Header
	TopicList    []*TopicsRow
	ForumList    []Forum
	DefaultForum int
	Sort         TopicListSort
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
	ItemList     []*ReplyUser
	ProfileOwner User
	CurrentScore int
	NextScore    int
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

type Account struct {
	*Header
	HTMLID   string
	TmplName string
	Inner    nobreak
}

type EmailListPage struct {
	*Header
	ItemList []Email
}

type AccountLoginsPage struct {
	*Header
	ItemList []LoginLogItem
	Paginator
}

type AccountDashPage struct {
	*Header
	MFASetup     bool
	CurrentScore int
	NextScore    int
	NextLevel    int
	Percentage   int
}

type LevelListItem struct {
	Level      int
	Score      int
	Status     string
	Percentage int // 0 to 200 to fit with the CSS logic
}

type LevelListPage struct {
	*Header
	Levels []LevelListItem
}

type ResetPage struct {
	*Header
	UID   int
	Token string
	MFA   bool
}

type ConvoListPage struct {
	*Header
	Convos []*Conversation
	Paginator
}

type ConvoViewRow struct {
	*ConversationPost
	User *User
	ClassName string
	ContentLines string
}

type ConvoViewPage struct {
	*Header
	Convo *Conversation
	Posts []ConvoViewRow
	CanModify bool
	Paginator
}

type ConvoCreatePage struct {
	*Header
	RecpName string
}

/* WIP for dyntmpl */
type Panel struct {
	*BasePanelPage
	HTMLID     string
	ClassNames string
	TmplName   string
	Inner      nobreak
}
type PanelAnalytics struct {
	*BasePanelPage
	FormAction string
	TmplName   string
	Inner      nobreak
}
type PanelAnalyticsStd struct {
	Graph     PanelTimeGraph
	ViewItems []PanelAnalyticsItem
	TimeRange string
	Unit      string
	TimeType  string
}
type PanelAnalyticsStdUnit struct {
	Graph     PanelTimeGraph
	ViewItems []PanelAnalyticsItemUnit
	TimeRange string
	Unit      string
	TimeType  string
}
type PanelAnalyticsActiveMemory struct {
	Graph     PanelTimeGraph
	ViewItems []PanelAnalyticsItemUnit
	TimeRange string
	Unit      string
	TimeType  string
	MemType   int
}

type PanelStats struct {
	Users       int
	Groups      int
	Forums      int
	Pages       int
	Settings    int
	WordFilters int
	Themes      int
	Reports     int
}

type BasePanelPage struct {
	*Header
	Stats         PanelStats
	Zone          string
	ReportForumID int
}

type PanelPage struct {
	*BasePanelPage
	ItemList  []interface{}
	Something interface{}
}

type GridElement struct {
	ID         string
	Href       string
	Body       string
	Order      int // For future use
	Class      string
	Background string
	TextColour string
	Note       string
}

type DashGrids struct {
	Grid1 []GridElement
	Grid2 []GridElement
}

type PanelDashboardPage struct {
	*BasePanelPage
	Grids DashGrids
}

type PanelSetting struct {
	*Setting
	FriendlyName string
}

type PanelSettingPage struct {
	*BasePanelPage
	ItemList []OptionLabel
	Setting  *PanelSetting
}

type PanelCustomPagesPage struct {
	*BasePanelPage
	ItemList []*CustomPage
	Paginator
}

type PanelCustomPageEditPage struct {
	*BasePanelPage
	Page *CustomPage
}

/*type PanelTimeGraph struct {
	Series []int64 // The counts on the left
	Labels []int64 // unixtimes for the bottom, gets converted into 1:00, 2:00, etc. with JS
}*/
type PanelTimeGraph struct {
	Series  [][]int64 // The counts on the left
	Labels  []int64   // unixtimes for the bottom, gets converted into 1:00, 2:00, etc. with JS
	Legends []string
}

type PanelAnalyticsItem struct {
	Time  int64
	Count int64
}
type PanelAnalyticsItemUnit struct {
	Time  int64
	Count int64
	Unit  string
}

type PanelAnalyticsPage struct {
	*BasePanelPage
	Graph     PanelTimeGraph
	ViewItems []PanelAnalyticsItem
	TimeRange string
	Unit      string
	TimeType  string
}

type PanelAnalyticsRoutesItem struct {
	Route string
	Count int
}

type PanelAnalyticsRoutesPage struct {
	*BasePanelPage
	ItemList  []PanelAnalyticsRoutesItem
	Graph     PanelTimeGraph
	TimeRange string
}

// TODO: Rename the fields as this structure is being used in a generic way now
type PanelAnalyticsAgentsItem struct {
	Agent         string
	FriendlyAgent string
	Count         int
}

type PanelAnalyticsAgentsPage struct {
	*BasePanelPage
	ItemList  []PanelAnalyticsAgentsItem
	TimeRange string
}

type PanelAnalyticsReferrersPage struct {
	*BasePanelPage
	ItemList  []PanelAnalyticsAgentsItem
	TimeRange string
	ShowSpam  bool
}

type PanelAnalyticsRoutePage struct {
	*BasePanelPage
	Route     string
	Graph     PanelTimeGraph
	ViewItems []PanelAnalyticsItem
	TimeRange string
}

type PanelAnalyticsAgentPage struct {
	*BasePanelPage
	Agent         string
	FriendlyAgent string
	Graph         PanelTimeGraph
	TimeRange     string
}

type PanelAnalyticsDuoPage struct {
	*BasePanelPage
	ItemList  []PanelAnalyticsAgentsItem
	Graph     PanelTimeGraph
	TimeRange string
}

type PanelThemesPage struct {
	*BasePanelPage
	PrimaryThemes []*Theme
	VariantThemes []*Theme
}

type PanelMenuListItem struct {
	Name      string
	ID        int
	ItemCount int
}

type PanelMenuListPage struct {
	*BasePanelPage
	ItemList []PanelMenuListItem
}

type PanelWidgetListPage struct {
	*BasePanelPage
	Docks       map[string][]WidgetEdit
	BlankWidget WidgetEdit
}

type PanelMenuPage struct {
	*BasePanelPage
	MenuID   int
	ItemList []MenuItem
}

type PanelMenuItemPage struct {
	*BasePanelPage
	Item MenuItem
}

type PanelUserPage struct {
	*BasePanelPage
	ItemList []*User
	Paginator
}

type PanelGroupPage struct {
	*BasePanelPage
	ItemList []GroupAdmin
	Paginator
}

type PanelEditGroupPage struct {
	*BasePanelPage
	ID          int
	Name        string
	Tag         string
	Rank        string
	DisableRank bool
}

type GroupForumPermPreset struct {
	Group         *Group
	Preset        string
	DefaultPreset bool
}

type PanelEditForumPage struct {
	*BasePanelPage
	ID     int
	Name   string
	Desc   string
	Active bool
	Preset string
	Groups []GroupForumPermPreset
}

type NameLangToggle struct {
	Name    string
	LangStr string
	Toggle  bool
}

type PanelEditForumGroupPage struct {
	*BasePanelPage
	ForumID int
	GroupID int
	Name    string
	Desc    string
	Active  bool
	Preset  string
	Perms   []NameLangToggle
}

type PanelEditGroupPermsPage struct {
	*BasePanelPage
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
	*BasePanelPage
	Backups []BackupItem
}

type PageLogItem struct {
	Action    template.HTML
	IPAddress string
	DoneAt    string
}

type PanelLogsPage struct {
	*BasePanelPage
	Logs []PageLogItem
	Paginator
}

type PageRegLogItem struct {
	RegLogItem
	ParsedReason string
}

type PanelRegLogsPage struct {
	*BasePanelPage
	Logs []PageRegLogItem
	Paginator
}

type DebugPageCache struct {
	Topics int
	Users int
	Replies int

	TCap int
	UCap int
	RCap int

	TopicListThaw bool
}

type DebugPageDatabase struct {
	Topics int
	Users int
	Replies int
	ProfileReplies int
	ActivityStream int
	Likes int
	Attachments int
	Polls int

	LoginLogs int
	RegLogs int
	ModLogs int
	AdminLogs int

	Views int
	ViewsAgents int
	ViewsForums int
	ViewsLangs int
	ViewsReferrers int
	ViewsSystems int
	PostChunks int
	TopicChunks int
}

type DebugPageDisk struct {
	Static int
	Attachments int
	Avatars int
	Logs int
	Backups int
	Git int
}

type PanelDebugPage struct {
	*BasePanelPage
	GoVersion string
	DBVersion string
	Uptime    string

	OpenConns int
	DBAdapter string

	Goroutines int
	CPUs       int
	MemStats   runtime.MemStats

	Cache DebugPageCache
	Database DebugPageDatabase
	Disk DebugPageDisk
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
func DefaultHeader(w http.ResponseWriter, user User) *Header {
	return &Header{Site: Site, Theme: Themes[fallbackTheme], CurrentUser: user, Writer: w}
}
func SimpleDefaultHeader(w http.ResponseWriter) *Header {
	return &Header{Site: Site, Theme: Themes[fallbackTheme], CurrentUser: GuestUser, Writer: w}
}
