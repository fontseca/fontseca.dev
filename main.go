package main

import (
  "context"
  "database/sql"
  "errors"
  "fmt"
  "fontseca.dev/handler"
  "fontseca.dev/repository"
  "fontseca.dev/service"
  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/binding"
  "github.com/go-playground/validator/v10"
  _ "github.com/lib/pq"
  "io"
  "log"
  "log/slog"
  "net/http"
  "os"
  "os/signal"
  "reflect"
  "strconv"
  "strings"
  "syscall"
  "time"
)

func main() {
  log.SetFlags(log.LstdFlags | log.Lshortfile)

  var db, err = sql.Open("postgres", fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s connect_timeout=5 sslmode=require binary_parameters=yes",
    mustLookupEnv("PG_USER"),
    mustLookupEnv("PG_PASSWORD"),
    mustLookupEnv("PG_HOST"),
    mustLookupEnv("PG_PORT"),
    mustLookupEnv("PG_DATABASE")))

  if nil != err {
    log.Fatal(err)
  }

  defer func(db *sql.DB) {
    fmt.Fprint(os.Stdout, "closing database... ")

    err := db.Close()
    if err != nil {
      fmt.Fprintf(os.Stderr, "could not close database: %v", err)
      return
    }

    fmt.Fprintln(os.Stdout, "done")
  }(db)

  if err = db.Ping(); nil != err {
    log.Fatal(err)
  }

  logfile, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
  if nil != err {
    log.Fatal(err)
  }

  defer logfile.Close()

  log.SetOutput(io.MultiWriter(os.Stderr, logfile))

  var mode = strings.TrimSpace(os.Getenv("SERVER_MODE"))
  if "" == mode {
    mode = gin.DebugMode
    slog.Warn("environment variable not found",
      slog.String("variable", "SERVER_MODE"),
      slog.String("default", mode))
  }

  gin.SetMode(mode)
  var engine = gin.New()

  engine.Use(gin.Recovery())
  engine.Use(func(c *gin.Context) {
  })

  var formatter = func(param gin.LogFormatterParams) string {
    if param.Latency > time.Minute {
      param.Latency = param.Latency.Truncate(time.Second)
    }

    bodySizeStr := "-"
    if param.BodySize > 0 {
      bodySizeStr = strconv.Itoa(param.BodySize)
    }

    // Logs messages with the Common Log Format.
    return fmt.Sprintf("%s - - [%s] \"%s %s %s\" %d %s in %s\n",
      param.ClientIP,
      param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
      param.Method,
      param.Path,
      param.Request.Proto,
      param.StatusCode,
      bodySizeStr,
      param.Latency,
    )
  }

  serverLogFile, err := os.OpenFile("server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
  if nil != err {
    slog.Error(err.Error())
    return
  }

  defer serverLogFile.Close()

  engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
    Formatter: formatter,
    Output:    serverLogFile,
  }))

  engine.Static("/public", "public")
  engine.StaticFile("/favicon.ico", "public/icons/favicon.ico")
  engine.StaticFile("/photo.webp", "public/images/photo.webp")

  binding.EnableDecoderDisallowUnknownFields = true
  if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
    v.RegisterTagNameFunc(func(fld reflect.StructField) string {
      var name = strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
      if 0 == strings.Compare(name, "-") {
        return ""
      }
      return name
    })
  }

  var (
    meRepository = repository.NewMeRepository(db)
    meService    = service.NewMeService(meRepository)
    me           = handler.NewMeHandler(meService)
  )

  meRepository.Register(context.Background())

  engine.GET("/me.info", me.Get)
  engine.POST("/me.setPhoto", me.SetPhoto)
  engine.POST("/me.setResume", me.SetResume)
  engine.POST("/me.setHireable", me.SetHireable)
  engine.POST("/me.set", me.Update)

  var (
    experienceRepository = repository.NewExperienceRepository(db)
    experienceService    = service.NewExperienceService(experienceRepository)
    experience           = handler.NewExperienceHandler(experienceService)
  )

  engine.GET("/me.experience.list", experience.Get)
  engine.GET("/me.experience.hidden.list", experience.GetHidden)
  engine.GET("/me.experience.info", experience.GetByID)
  engine.POST("/me.experience.add", experience.Add)
  engine.POST("/me.experience.set", experience.Set)
  engine.POST("/me.experience.hide", experience.Hide)
  engine.POST("/me.experience.show", experience.Show)
  engine.POST("/me.experience.quit", experience.Quit)
  engine.POST("/me.experience.remove", experience.Remove)

  var (
    technologyTagRepository = repository.NewTechnologyTagRepository(db)
    technologyTagService    = service.NewTechnologyTagService(technologyTagRepository)
    technologies            = handler.NewTechnologyTagHandler(technologyTagService)
  )

  engine.GET("/technologies.list", technologies.Get)
  engine.POST("/technologies.add", technologies.Add)
  engine.POST("/technologies.set", technologies.Set)
  engine.POST("/technologies.remove", technologies.Remove)

  var (
    projectsRepository = repository.NewProjectsRepository(db)
    projectsService    = service.NewProjectsService(projectsRepository)
    projects           = handler.NewProjectsHandler(projectsService)
  )

  engine.GET("/me.projects.list", projects.Get)
  engine.GET("/me.projects.info", projects.GetByID)
  engine.GET("/me.projects.archived.list", projects.GetArchived)
  engine.POST("/me.projects.add", projects.Add)
  engine.POST("/me.projects.set", projects.Set)
  engine.POST("/me.projects.archive", projects.Archive)
  engine.POST("/me.projects.unarchive", projects.Unarchive)
  engine.POST("/me.projects.finish", projects.Finish)
  engine.POST("/me.projects.unfinish", projects.Unfinish)
  engine.POST("/me.projects.remove", projects.Remove)
  engine.POST("/me.projects.setPlaygroundURL", projects.SetPlaygroundURL)
  engine.POST("/me.projects.setFirstImageURL", projects.SetFirstImageURL)
  engine.POST("/me.projects.setSecondImageURL", projects.SetSecondImageURL)
  engine.POST("/me.projects.setGitHubURL", projects.SetGitHubURL)
  engine.POST("/me.projects.setCollectionURL", projects.SetCollectionURL)
  engine.POST("/me.projects.technologies.add", projects.AddTechnologyTag)
  engine.POST("/me.projects.technologies.remove", projects.RemoveTechnologyTag)

  var archive = repository.NewArchiveRepository(db)

  var (
    tagsRepository = repository.NewTagsRepository(db)
    tagsService    = service.NewTagsService(tagsRepository)
    tags           = handler.NewTagsHandler(tagsService)
  )

  engine.POST("/archive.tags.add", tags.Add)
  engine.GET("/archive.tags.list", tags.Get)
  engine.POST("/archive.tags.set", tags.Update)
  engine.POST("/archive.tags.remove", tags.Remove)

  var (
    topicsRepository = repository.NewTopicsRepository(db)
    topicsService    = service.NewTopicsService(topicsRepository)
    topics           = handler.NewTopicsHandler(topicsService)
  )

  engine.POST("/archive.topics.add", topics.Add)
  engine.GET("/archive.topics.list", topics.Get)
  engine.POST("/archive.topics.set", topics.Update)
  engine.POST("/archive.topics.remove", topics.Remove)

  var (
    draftsService = service.NewDraftsService(archive)
    drafts        = handler.NewDraftsHandler(draftsService)
  )

  engine.POST("/archive.drafts.start", drafts.Start)
  engine.POST("/archive.drafts.publish", drafts.Publish)
  engine.GET("/archive.drafts.list", drafts.Get)
  engine.GET("/archive.drafts.info", drafts.GetByID)
  engine.POST("/archive.drafts.share", drafts.Share)
  engine.POST("/archive.drafts.revise", drafts.Revise)
  engine.POST("/archive.drafts.discard", drafts.Discard)
  engine.POST("/archive.drafts.tags.add", drafts.AddTag)
  engine.POST("/archive.drafts.tags.remove", drafts.RemoveTag)

  var (
    articlesService = service.NewArticlesService(archive)
    articles        = handler.NewArticlesHandler(articlesService)
  )

  engine.GET("/archive.articles.list", articles.Get)
  engine.GET("/archive.articles.hidden.list", articles.GetHidden)
  engine.GET("/archive.articles.info", articles.GetByID)
  engine.POST("/archive.articles.amend", articles.Amend)
  engine.POST("/archive.articles.setSlug", articles.SetSlug)
  engine.POST("/archive.articles.hide", articles.Hide)
  engine.POST("/archive.articles.show", articles.Show)
  engine.POST("/archive.articles.remove", articles.Remove)
  engine.POST("/archive.articles.pin", articles.Pin)
  engine.POST("/archive.articles.unpin", articles.Unpin)
  engine.POST("/archive.articles.tags.add", articles.AddTag)
  engine.POST("/archive.articles.tags.remove", articles.RemoveTag)

  var (
    patchesServices = service.NewPatchesService(archive)
    patches         = handler.NewPatchesHandler(patchesServices)
  )

  engine.GET("/archive.articles.patches.list", patches.Get)
  engine.POST("/archive.articles.patches.revise", patches.Revise)
  engine.POST("/archive.articles.patches.share", patches.Share)
  engine.POST("/archive.articles.patches.discard", patches.Discard)
  engine.POST("/archive.articles.patches.release", patches.Release)

  var web = handler.NewWebHandler(
    meService,
    experienceService,
    projectsService,
    draftsService,
    articlesService,
    topicsService,
    tagsService,
  )

  engine.GET("/", web.RenderMe)
  engine.GET("/experience", web.RenderExperience)
  engine.GET("/work", web.RenderProjects)
  engine.GET("/work/:project_slug", web.RenderProjectDetails)
  engine.GET("/archive", web.RenderArchive)
  engine.GET("/archive/:topic", web.RenderArchive)
  engine.GET("/archive/:topic/:year/:month", web.RenderArchive)
  engine.GET("/archive/tag/:tag", web.RenderArchive)
  engine.GET("/archive/:topic/:year/:month/:slug", web.RenderArticle)
  engine.GET("/archive/sharing/:hash", web.RenderArticle)

  engine.NoRoute(func(c *gin.Context) {
    http.Error(c.Writer, "404 Not Found", http.StatusNotFound)
  })

  engine.HandleMethodNotAllowed = true
  var routes = engine.Routes()
  engine.NoMethod(func(c *gin.Context) {
    if "Bearer" != strings.Split(strings.TrimSpace(c.GetHeader("Authorization")), " ")[0] {
      http.Error(c.Writer, "404 Not Found", http.StatusNotFound)
      return
    }

    var allowedMethods = make([]string, 0, 1)
    for _, route := range routes {
      if route.Path == c.Request.URL.Path {
        allowedMethods = append(allowedMethods, route.Method)
      }
    }

    c.Header("Allow", strings.Join(allowedMethods, ","))
    http.Error(c.Writer, "405 Not Allowed", http.StatusMethodNotAllowed)
  })

  if gin.ReleaseMode != mode {
    emitWelcome()
  }

  var port = strings.TrimSpace(os.Getenv("PORT"))
  if "" == port {
    port = "8080"
    slog.Warn("environment variable not found",
      slog.String("variable", "PORT"),
      slog.String("default", port))
  }

  var server = http.Server{
    Addr:           "0.0.0.0:" + port,
    IdleTimeout:    1 * time.Minute,
    ReadTimeout:    5 * time.Second,
    WriteTimeout:   5 * time.Second,
    MaxHeaderBytes: 1024,
    Handler:        engine,
  }

  slog.Info("running server",
    slog.String("address", server.Addr),
    slog.String("mode", mode))

  var (
    didNotServe = make(chan struct{})
    shutdown    = make(chan os.Signal, 1)
  )

  signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

  go func() {
    err = server.ListenAndServe()
    if !errors.Is(err, http.ErrServerClosed) {
      slog.Error(err.Error())
    }

    didNotServe <- struct{}{}
  }()

  select {
  case <-didNotServe:
    return
  case sig := <-shutdown:
    fmt.Fprintf(os.Stdout, "received %s signal, gracefully shutting down...\n", sig.String())

    archive.Close()

    if err := server.Shutdown(context.TODO()); nil != err {
      fmt.Fprintf(os.Stderr, "could not shutdown server: %v", err)
    }
  }
}

func mustLookupEnv(key string) string {
  v, ok := os.LookupEnv(key)

  if !ok {
    log.Fatalf("could not load environment variable '%s'", key)
  }

  return v
}

func emitWelcome() {
  lines := []string{
    "\x1B[0m\x1B[1;96m ________  ________   ________    _________   ________   _______    ________   ________        \x1B[1;91m________   _______    ___      ___ \x1B[0m\n",
    "\x1B[0m\x1B[1;96m|\\  _____\\|\\   __  \\ |\\   ___  \\ |\\___   ___\\|\\   ____\\ |\\  ___ \\  |\\   ____\\ |\\   __  \\      \x1B[1;91m|\\   ___ \\ |\\  ___ \\  |\\  \\    /  /|\x1B[0m\n",
    "\x1B[0m\x1B[1;96m\\ \\  \\__/ \\ \\  \\|\\  \\\\ \\  \\\\ \\  \\\\|___ \\  \\_|\\ \\  \\___|_\\ \\   __/| \\ \\  \\___| \\ \\  \\|\\  \\     \x1B[1;91m\\ \\  \\_|\\ \\\\ \\   __/| \\ \\  \\  /  / /\x1B[0m\n",
    "\x1B[0m\x1B[1;96m \\ \\   __\\ \\ \\  \\\\\\  \\\\ \\  \\\\ \\  \\    \\ \\  \\  \\ \\_____  \\\\ \\  \\_|/__\\ \\  \\     \\ \\   __  \\     \x1B[1;91m\\ \\  \\ \\\\ \\\\ \\  \\_|/__\\ \\  \\/  / / \x1B[0m\n",
    "\x1B[0m\x1B[1;96m  \\ \\  \\_|  \\ \\  \\\\\\  \\\\ \\  \\\\ \\  \\    \\ \\  \\  \\|____|\\  \\\\ \\  \\_|\\ \\\\ \\  \\____ \\ \\  \\ \\  \\  \x1B[1;91m___\x1B[1;91m\\ \\  \\_\\\\ \\\\ \\  \\_|\\ \\\\ \\    / /  \x1B[0m\n",
    "\x1B[0m\x1B[1;96m   \\ \\__\\    \\ \\_______\\\\ \\__\\\\ \\__\\    \\ \\__\\   ____\\_\\  \\\\ \\_______\\\\ \\_______\\\\ \\__\\ \\__\\\x1B[1;91m|\\__\\\\ \\_______\\\\ \\_______\\\\ \\__/ /   \x1B[0m\n",
    "\x1B[0m\x1B[1;96m    \\|__|     \\|_______| \\|__| \\|__|     \\|__|  |\\_________\\\\|_______| \\|_______| \\|__|\\|__|\x1B[1;91m\\|__| \\|_______| \\|_______| \\|__|/    \x1B[0m\n",
    "\x1B[0m\x1B[1;96m                                                \\\\|_________|                                                                     \x1B[0m\n",
    "\x1B[0m                                                  Make it simple. Make it possible.                                               \n",
    "\x1B[0m                                              https://github.com/fontseca/fontseca.dev                                            \n\n"}

  factor := 10

  for _, line := range lines {
    time.Sleep(time.Duration(factor) * time.Millisecond)
    fmt.Print(line)
  }
}
