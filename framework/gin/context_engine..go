package gin

const (
	// PlatformGoogleAppEngine when running on Google App Engine. Trust X-Appengine-Remote-Addr
	// for determining the client's IP
	PlatformGoogleAppEngine = "X-Appengine-Remote-Addr"
	// PlatformCloudflare when using Cloudflare's CDN. Trust CF-Connecting-IP for determining
	// the client's IP
	PlatformCloudflare = "CF-Connecting-IP"
	// PlatformFlyIO when running on Fly.io. Trust Fly-Client-IP for determining the client's IP
	PlatformFlyIO = "Fly-Client-IP"
)

func init() {
	defaultPlatform = PlatformGoogleAppEngine
}
