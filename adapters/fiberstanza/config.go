package fiberstanza

// Config defines the config for fiberstanza middleware.
type Config struct {
	ResourceName string `json:"resourceName"` // optional (but required if you want to protect multiple resources)
}
