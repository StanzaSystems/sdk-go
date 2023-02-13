package fiberstanza

import (
	"context"
	"fmt"
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
	APIKey string // customer generated API key

	// Optional
	Name        string // defines applications name
	Release     string // defines applications version
	Environment string // defines applications environment
	StanzaHub   string // host:port (ipv4, ipv6, or resolvable hostname)
	DataSource  string // local:<path>, consul:<key>, or grpc:host:port
}

// Decorator defines the config for fiberstanza middleware.
type Decorator struct {
	Name string // optional (but required if you want to use multiple Decorators)
}

// New creates a new fiberstanza middleware fiber.Handler
func New(d Decorator) fiber.Handler {
	if err := stanza.NewDecorator(d.Name); err != nil {
		logging.Error(fmt.Errorf("failed to register new decorator: %v", err))
	}
	im, err := stanza.InitHttpInboundMeters(d.Name)
	if err != nil {
		logging.Error(fmt.Errorf("failed to initialize new http inbound meters: %v", err))
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
			logging.Error(fmt.Errorf("failed to convert request from fasthttp: %v", err))
			return c.Next() // log error and fail open
		}
		ctx, status := stanza.HttpInboundHandler(savedCtx, d.Name, c.Route().Path, &im, &req)
		if status != http.StatusOK {
			return c.SendStatus(status)
		}
		c.SetUserContext(ctx)
		return c.Next()
	}
}

func Init(ctx context.Context, client Client) (func(), error) {
	return stanza.Init(ctx, stanza.ClientOptions(client))
}
