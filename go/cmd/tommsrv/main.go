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
			return next(&myContext{c, &storage.SimpleProbSpecStore{DB: db}, &storage.VolatileProbResultStore{}})
		}
	}
}

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
	dashboardAPI.GET("/probe/results/3dforce", getAllResults3DForceGraph)
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
			Root:    "./dist-ng",
		}))
	agent := e.Group("/agent")
	agent.Use(middleware.StaticWithConfig(
		middleware.StaticConfig{
			Skipper: middleware.DefaultSkipper,
			Index:   "index.html",
			HTML5:   false,
			Root:    "./dist-agent",
		}))

	threedforce := e.Group("/3dforce")
	threedforce.Use(middleware.StaticWithConfig(
		middleware.StaticConfig{
			Skipper: middleware.DefaultSkipper,
			Index:   "index.html",
			HTML5:   false,
			Root:    "./3dforce",
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
	var filter = c.QueryParam("filter")
	res, err := ctx.probeResultStore.GetResultsWithSubstring(ctx.Context.Request().Context(), filter)
	if err != nil {
		return err
	}
	c.JSON(200, res)
	return nil
}

type Node struct {
	ID    string `json:"id"`
	Group string `json:"group"`
}

type Link struct {
	Source  string `json:"source"`
	Target  string `json:"target"`
	Status  string `json:"status"`
	Elapsed int    `json:"elapsed"`
	Port    string `json:"port"`
	Comment string `json:"comment"`
}

func getAllResults3DForceGraph(c echo.Context) error {
	var ctx = c.(*myContext)
	var err error
	var filter = c.QueryParam("filter")
	res, err := ctx.probeResultStore.GetResultsWithSubstring(ctx.Context.Request().Context(), filter)
	if err != nil {
		return err
	}
	var nodes = make([]Node, 0, 500)
	var links = make([]Link, 0, 500)

	var nodeMap = map[string]*Node{}

	for _, r := range res {
		if r.Type != "tcp" || len(r.Args) < 2 {
			continue
		}
		var source = r.Source
		var target = r.Args[0]
		var port = r.Args[1]

		if v, ok := nodeMap[source]; ok {
			if v.Group == "target" {
				v.Group = "source+target"
			}
		} else {
			nodeMap[source] = &Node{ID: source, Group: "source"}
		}
		if v, ok := nodeMap[target]; ok {
			if v.Group == "source" {
				v.Group = "source+target"
			}
		} else {
			nodeMap[target] = &Node{ID: target, Group: "target"}
		}
		links = append(links, Link{Source: source, Target: target, Status: r.Status, Elapsed: r.Elapsed, Port: port, Comment: r.Comment})
	}

	for _, v := range nodeMap {
		nodes = append(nodes, *v)
	}

	c.JSON(200, map[string]interface{}{
		"nodes": nodes,
		"links": links,
	})
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
	err = ctx.probeSpecStore.PutStoredProbeSpecRule(ctx.Context.Request().Context(), id, nil)
	if err != nil {
		return err
	}
	c.JSON(http.StatusNoContent, nil)
	return nil
}
