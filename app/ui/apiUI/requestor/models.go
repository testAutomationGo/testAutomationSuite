package requestor

type SavedRequest struct {
	Name                string
	URL                 string
	RequestType         string
	Headers             map[string]string
	Body                string
	AuthenticationType  string
	AuthenticationToken string
}
