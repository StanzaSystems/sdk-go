package fiberstanza

import (
	"context"
	"net/http"
	"time"

	"github.com/StanzaSystems/sdk-go/logging"
	"github.com/StanzaSystems/sdk-go/stanza"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Client defines options for a new Stanza Client
type Client struct {
	// Required
	// DSN or some other kind of customer key/ID

	// Optional
	Name        string // defines applications unique name
	Release     string // defines applications version
	Environment string // defines applications environment
	StanzaHub   string // host:port (ipv4, ipv6, or resolveable hostname)
	DataSource  string // local:<path>, consul:<key>, or grpc:host:port
}

// Decorator defines the config for fiberstanza middleware.
type Decorator struct {
	Name string // optional (but required if you want to use multiple Decorators)
}

// New creates a new fiberstanza middleware fiber.Handler
func New(d Decorator) fiber.Handler {
	if err := stanza.NewResource(d.Name); err != nil {
		logging.Error(err, "failed to register new resource")
	}
	im, err := stanza.InitHttpInboundMeters(d.Name)
	if err != nil {
		logging.Error(err, "failed to initialize new http inbound meters")
	}

	return func(c *fiber.Ctx) error {
		start := time.Now()
		savedCtx, cancel := context.WithCancel(c.UserContext())

		im.ActiveRequests.Add(savedCtx, 1, im.Attributes...)
		defer func() {
			im.Duration.Record(savedCtx, float64(time.Since(start).Microseconds())/1000, im.Attributes...)
			im.RequestSize.Record(savedCtx, int64(len(c.Request().Body())), im.Attributes...)
			im.ResponseSize.Record(savedCtx, int64(len(c.Response().Body())), im.Attributes...)
			im.ActiveRequests.Add(savedCtx, -1, im.Attributes...)
			c.SetUserContext(savedCtx)
			cancel()
		}()

		// TODO(msg): implement HttpInboundHandler as fasthttp handler instead of converting to net/http?
		var req http.Request
		if err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true); err != nil {
			logging.Error(err, "failed to convert request from fasthttp")
			return c.Next() // log error and fail open
		}
		ctx, status := stanza.HttpInboundHandler(savedCtx, c.Route().Path, &im, &req)
		if status != http.StatusOK {
			return c.SendStatus(status)
		}
		c.SetUserContext(ctx)
		return c.Next()
	}
}

func Init(ctx context.Context, client Client) error {
	return stanza.Init(ctx, stanza.Client(client))
}
