package alerts

// TODO: Move the other alert related stuff to package alerts, maybe move notification logic here too?

type AlertItem struct {
	ASID    int
	Path    string
	Message string
	Avatar  string
}
