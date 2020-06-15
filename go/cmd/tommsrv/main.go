package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"mxmz.it/mxmz/tommaso/dto"
	"mxmz.it/mxmz/tommaso/ports"
	"mxmz.it/mxmz/tommaso/storage"
)

func init() {

}

type myContext struct {
	echo.Context
	probeSpecStore   ports.ProbeSpecStore
	probeResultStore ports.ProbeResultStore
}

func setupContext(db *storage.FileDB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(&myContext{c, &storage.SimpleProbSpecStore{db}, &storage.VolatileProbResultStore{}})
		}
	}
}

func StoredProbeSpecFactory() interface{}     { return new(dto.StoredProbeSpec) }
func StoredProbeSpecRuleFactory() interface{} { return new(dto.StoredProbeSpecRule) }

func main() {

	var db = storage.DefaultFileDB()

	e := echo.New()
	agentAPI := e.Group("/api/agent")
	agentAPI.Use(setupContext(db))
	agentAPI.POST("/get-my-probe-specs", getMyProbeSpecs)
	agentAPI.POST("/push-my-probe-results", pushMyProbeResults)
	dashboardAPI := e.Group("/api/dashboard")
	dashboardAPI.Use(setupContext(db))
	dashboardAPI.GET("/probe/results", getAllProbeResults)
	dashboardAPI.GET("/probe/specs", listProbeSpecs)
	dashboardAPI.PUT("/probe/specs/:id", putProbeSpec)
	dashboardAPI.DELETE("/probe/specs/:id", deleteProbeSpec)
	dashboardAPI.GET("/probe/rules", listProbeSpecRules)
	dashboardAPI.PUT("/probe/rules/:id", putProbeSpecRule)
	dashboardAPI.DELETE("/probe/rules/:id", deleteProbeSpecRule)
	e.Use(middleware.StaticWithConfig(
		middleware.StaticConfig{
			Skipper: middleware.DefaultSkipper,
			Index:   "index.html",
			HTML5:   true,
			Root:    "./dist",
		}))

	// Start server

	e.Logger.Fatal(e.Start(":7997"))
}

func getMyProbeSpecs(c echo.Context) error {
	var ctx = c.(*myContext)
	var err error
	var mySrcs dto.MySources
	if err = c.Bind(&mySrcs); err != nil {
		return err
	}
	probSpecs, err := ctx.probeSpecStore.GetProbeSpecsForNames(ctx.Context.Request().Context(), &mySrcs)
	if err != nil {
		return err
	}
	c.JSON(200, probSpecs)
	return nil
}
func pushMyProbeResults(c echo.Context) error {
	var ctx = c.(*myContext)
	var err error
	var myResults []*dto.ProbeResult
	if err = c.Bind(&myResults); err != nil {
		return err
	}
	err = ctx.probeResultStore.PutResultsForSources(ctx.Context.Request().Context(), myResults)
	if err != nil {
		return err
	}
	c.JSON(http.StatusAccepted, nil)
	return nil
}

func getAllProbeResults(c echo.Context) error {
	var ctx = c.(*myContext)
	var err error
	var myResults []*dto.ProbeResult
	if err = c.Bind(&myResults); err != nil {
		return err
	}
	res, err := ctx.probeResultStore.GetResultsBySourcePrefix(ctx.Context.Request().Context(), "")
	if err != nil {
		return err
	}
	c.JSON(200, res)
	return nil
}

func listProbeSpecs(c echo.Context) error {
	var ctx = c.(*myContext)
	var err error
	probSpecs, err := ctx.probeSpecStore.GetStoredProbeSpecs(ctx.Context.Request().Context())
	if err != nil {
		return err
	}
	c.JSON(200, probSpecs)
	return nil
}

func putProbeSpec(c echo.Context) error {
	var ctx = c.(*myContext)
	var id = c.Param("id")
	var err error
	var spec dto.ProbeSpec
	if err = c.Bind(&spec); err != nil {
		return err
	}

	err = ctx.probeSpecStore.PutStoredProbeSpec(ctx.Context.Request().Context(), id, &spec)
	if err != nil {
		return err
	}
	c.JSON(http.StatusNoContent, nil)
	return nil
}
func deleteProbeSpec(c echo.Context) error {
	var ctx = c.(*myContext)
	var id = c.Param("id")
	var err error
	err = ctx.probeSpecStore.PutStoredProbeSpec(ctx.Context.Request().Context(), id, nil)
	if err != nil {
		return err
	}
	c.JSON(http.StatusNoContent, nil)
	return nil
}

func listProbeSpecRules(c echo.Context) error {
	var ctx = c.(*myContext)
	var err error
	probSpecRules, err := ctx.probeSpecStore.GetStoredProbeSpecRules(ctx.Context.Request().Context())
	if err != nil {
		return err
	}
	c.JSON(200, probSpecRules)
	return nil
}

func putProbeSpecRule(c echo.Context) error {
	var ctx = c.(*myContext)
	var id = c.Param("id")
	var err error
	var rule dto.ProbeSpecRule
	if err = c.Bind(&rule); err != nil {
		return err
	}

	err = ctx.probeSpecStore.PutStoredProbeSpecRule(ctx.Context.Request().Context(), id, &rule)
	if err != nil {
		return err
	}
	c.JSON(http.StatusNoContent, nil)
	return nil
}
func deleteProbeSpecRule(c echo.Context) error {
	var ctx = c.(*myContext)
	var id = c.Param("id")
	var err error
	err = ctx.probeSpecStore.PutStoredProbeSpec(ctx.Context.Request().Context(), id, nil)
	if err != nil {
		return err
	}
	c.JSON(http.StatusNoContent, nil)
	return nil
}
