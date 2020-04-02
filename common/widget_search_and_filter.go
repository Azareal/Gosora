package common

import "errors"

// TODO: Move this into it's own package to make neater and tidier
type filterForum struct {
	*Forum
	Selected bool
}
type searchAndFilter struct {
	*Header
	Forums []filterForum
}

func widgetSearchAndFilter(widget *Widget, hvars interface{}) (out string, err error) {
	header := hvars.(*Header)
	u := header.CurrentUser
	var forums []filterForum
	var canSee []int
	if u.IsSuperAdmin {
		canSee, err = Forums.GetAllVisibleIDs()
		if err != nil {
			return "", err
		}
	} else {
		group, err := Groups.Get(u.Group)
		if err != nil {
			// TODO: Revisit this
			return "", errors.New("Something weird happened")
		}
		canSee = group.CanSee
	}

	for _, fid := range canSee {
		f := Forums.DirtyGet(fid)
		if f.ParentID == 0 && f.Name != "" && f.Active {
			forums = append(forums, filterForum{f, (header.Zone == "view_forum" || header.Zone == "topics") && header.ZoneID == f.ID})
		}
	}

	saf := &searchAndFilter{header, forums}
	err = saf.Header.Theme.RunTmpl("widget_search_and_filter", saf, saf.Header.Writer)
	return "", err
}
