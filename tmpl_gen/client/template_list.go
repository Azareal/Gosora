package tmpl

var alert_frags = make([][]byte,9)

// nolint
func init() {
alert_frags[0] = []byte(`<div class='alertItem withAvatar' style='background-image:url("`)
alert_frags[1] = []byte(`");'><a class='text' data-asid='`)
alert_frags[2] = []byte(`' href="`)
alert_frags[3] = []byte(`">`)
alert_frags[4] = []byte(`</a></div>`)
alert_frags[5] = []byte(`
<div class='alertItem'><a href="`)
alert_frags[6] = []byte(`" class='text'>`)
alert_frags[7] = []byte(`</a></div>`)
}
