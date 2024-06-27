package main

import (
  "context"
  "database/sql"
  "errors"
  "fmt"
  "fontseca.dev/handler"
  "fontseca.dev/problem"
  "fontseca.dev/repository"
  "fontseca.dev/service"
  "github.com/gin-gonic/gin"
  "github.com/gin-gonic/gin/binding"
  "github.com/go-playground/validator/v10"
  "github.com/google/uuid"
  "github.com/mattn/go-sqlite3"
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

// table contains information about a relation in the database.
type table struct {
  name       string
  definition string
}

// exists checks if the table t.name is already created in the transaction tx.
func (t *table) exists(ctx context.Context, tx *sql.Tx) bool {
  if nil == tx {
    return false
  }
  var query = `
  SELECT count (1)
    FROM "sqlite_master"
   WHERE "type" = 'table'
     AND "name" = $1;`
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  var result = tx.QueryRowContext(ctx, query, t.name)
  var err = result.Err()
  if nil != err {
    err = fmt.Errorf("checking existence of table %q: %v", t.name, err)
    if rollbackErr := tx.Rollback(); nil != rollbackErr {
      log.Fatalf("unable to rollback: %v: %v", err, rollbackErr)
    }
    log.Fatal(err)
  }
  var n int
  err = result.Scan(&n)
  if nil != err {
    log.Fatal(err)
  }
  return n >= 1
}

// create attempts to create the table t in the transaction tx.
func (t *table) create(ctx context.Context, tx *sql.Tx) {
  if nil == tx {
    return
  }
  ctx, cancel := context.WithTimeout(ctx, time.Second)
  defer cancel()
  if _, err := tx.ExecContext(ctx, t.definition); nil != err {
    err = fmt.Errorf("creating table %q: %v", t.name, err)
    if rollbackErr := tx.Rollback(); nil != rollbackErr {
      log.Fatalf("unable to rollback: %v: %v", err, rollbackErr)
    }
    log.Fatal(err)
  }
}

func main() {
  log.SetFlags(log.LstdFlags | log.Lshortfile)

  sql.Register("sqlite3_custom", &sqlite3.SQLiteDriver{
    ConnectHook: func(conn *sqlite3.SQLiteConn) error {
      if err := conn.RegisterFunc(
        "uuid_generate_v4",
        func() string { return uuid.New().String() },
        true,
      ); nil != err {
        return err
      }

      if err := conn.RegisterFunc(
        "uuid_nil",
        func() string { return uuid.Nil.String() },
        true,
      ); nil != err {
        return err
      }

      return nil
    },
  })

  var db, err = sql.Open("sqlite3_custom", "./db.sqlite")
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

  var tables = []table{
    {
      name: "me",
      definition: `
      CREATE TABLE "me"
      (
        "username"      VARCHAR(64) UNIQUE NOT NULL DEFAULT 'fontseca.dev',
        "first_name"    VARCHAR(6) NOT NULL DEFAULT 'Jeremy',
        "last_name"     VARCHAR(7) NOT NULL DEFAULT 'Fonseca',
        "summary"       VARCHAR(1024) NOT NULL,
        "job_title"     VARCHAR(64) NOT NULL DEFAULT 'Back-End Software Developer',
        "email"         VARCHAR(254) NOT NULL,
        "photo_url"     VARCHAR(2048) NOT NULL DEFAULT 'about:blank',
        "resume_url"    VARCHAR(2048) NOT NULL DEFAULT 'about:blank',
        "coding_since"  INT NOT NULL DEFAULT 2017,
        "company"       VARCHAR(64),
        "location"      VARCHAR(64),
        "hireable"      BOOLEAN NOT NULL DEFAULT TRUE,
        "github_url"    VARCHAR(2048) NOT NULL DEFAULT 'https://github.com/fontseca',
        "linkedin_url"  VARCHAR(2048) NOT NULL DEFAULT 'about:blank',
        "youtube_url"   VARCHAR(2048) NOT NULL DEFAULT 'about:blank',
        "twitter_url"   VARCHAR(2048) NOT NULL DEFAULT 'about:blank',
        "instagram_url" VARCHAR(2048) NOT NULL DEFAULT 'about:blank',
        "created_at"    TIMESTAMP NOT NULL DEFAULT current_timestamp,
        "updated_at"    TIMESTAMP NOT NULL DEFAULT current_timestamp,
        CHECK ("coding_since" = 2017)
      );`,
    },
    {
      name: "experience",
      definition: `
      CREATE TABLE "experience"
      (
        "uuid"       VARCHAR(36) NOT NULL PRIMARY KEY DEFAULT (uuid_generate_v4 ()),
        "starts"     INT NOT NULL,
        "ends"       INT NULL,
        "job_title"  VARCHAR(64) NOT NULL DEFAULT 'Back-End Software Developer',
        "company"    VARCHAR(64) NOT NULL,
        "country"    VARCHAR(64),
        "summary"    TEXT NOT NULL,
        "active"     BOOLEAN DEFAULT FALSE,
        "hidden"     BOOLEAN DEFAULT FALSE,
        "created_at" TIMESTAMP NOT NULL DEFAULT current_timestamp,
        "updated_at" TIMESTAMP NOT NULL DEFAULT current_timestamp,
        CHECK ("starts" > 2017),
        CHECK ("ends" > 2017 OR "ends" IS NULL)
      );`,
    },
    {
      name: "project",
      definition: `
      CREATE TABLE "project"
      (
        "uuid"             VARCHAR(36) NOT NULL PRIMARY KEY DEFAULT (uuid_generate_v4 ()),
        "name"             VARCHAR(64) NOT NULL,
        "slug"             VARCHAR(2024) NOT NULL,
        "homepage"         VARCHAR(2048) NOT NULL ON CONFLICT REPLACE DEFAULT 'about:blank',
        "language"         VARCHAR(64) NULL,
        "summary"          VARCHAR(1024) NOT NULL ON CONFLICT REPLACE DEFAULT 'No summary.',
        "read_time"        INT NOT NULL ON CONFLICT REPLACE DEFAULT 0,
        "content"          TEXT NOT NULL ON CONFLICT REPLACE DEFAULT 'No content.',
        "estimated_time"   INT DEFAULT NULL,
        "first_image_url"  VARCHAR(2048) NOT NULL ON CONFLICT REPLACE DEFAULT 'about:blank',
        "second_image_url" VARCHAR(2048) NOT NULL ON CONFLICT REPLACE DEFAULT 'about:blank',
        "github_url"       VARCHAR(2048) NOT NULL ON CONFLICT REPLACE DEFAULT 'about:blank',
        "collection_url"   VARCHAR(2048) NOT NULL ON CONFLICT REPLACE DEFAULT 'about:blank',
        "playground_url"   VARCHAR(2048) NOT NULL ON CONFLICT REPLACE DEFAULT 'about:blank',
        "playable"         BOOLEAN NOT NULL DEFAULT FALSE,
        "archived"         BOOLEAN NOT NULL DEFAULT FALSE,
        "finished"         BOOLEAN DEFAULT FALSE,
        "created_at"       TIMESTAMP NOT NULL DEFAULT current_timestamp,
        "updated_at"       TIMESTAMP NOT NULL DEFAULT current_timestamp
      );`,
    },
    {
      name: "technology_tag",
      definition: `
      CREATE TABLE "technology_tag"
      (
        "uuid"        VARCHAR(36) NOT NULL PRIMARY KEY DEFAULT (uuid_generate_v4 ()),
        "name"        VARCHAR(64) NOT NULL,
        "created_at"  TIMESTAMP NOT NULL DEFAULT current_timestamp,
        "updated_at"  TIMESTAMP NOT NULL DEFAULT current_timestamp
      );`,
    },
    {
      name: "project_technology_tag",
      definition: `
      CREATE TABLE "project_technology_tag"
      (
        "project_uuid"        VARCHAR(36) NOT NULL REFERENCES "project" ("uuid"),
        "technology_tag_uuid" VARCHAR(36) NOT NULL REFERENCES "technology_tag" ("uuid")
      );`,
    },
    {
      name: "topic",
      definition: `
      CREATE TABLE "topic"
      (
        "id"         VARCHAR(32) PRIMARY KEY,
        "name"       VARCHAR(32) NOT NULL,
        "created_at" TIMESTAMP NOT NULL DEFAULT current_timestamp,
        "updated_at" TIMESTAMP NOT NULL DEFAULT current_timestamp
      );`,
    },
    {
      name: "article",
      definition: `
      CREATE TABLE "article"
      (
        "uuid"         VARCHAR(36) NOT NULL PRIMARY KEY DEFAULT (uuid_generate_v4 ()),
        "title"        VARCHAR(256) NOT NULL,
        "author"       VARCHAR(64) NOT NULL REFERENCES "me" ("username"),
        "slug"         VARCHAR(512) NOT NULL,
        "read_time"    INT NOT NULL ON CONFLICT REPLACE DEFAULT 0,
        "views"        INTEGER NOT NULL ON CONFLICT REPLACE DEFAULT 0,
        "content"      TEXT NOT NULL ON CONFLICT REPLACE DEFAULT 'No content.',
        "draft"        BOOLEAN DEFAULT TRUE,
        "pinned"       BOOLEAN DEFAULT FALSE,
        "hidden"       BOOLEAN DEFAULT FALSE,
        "topic"        VARCHAR(32) REFERENCES "topic" ("id")
        "drafted_at"   TIMESTAMP NOT NULL DEFAULT current_timestamp,
        "published_at" TIMESTAMP DEFAULT NULL,
        "updated_at"   TIMESTAMP NOT NULL DEFAULT current_timestamp
        "modified_at"  TIMESTAMP DEFAULT NULL,
      );`,
    },
    {
      name: "article_patch",
      definition: `
      CREATE TABLE "article_patch"
      (
        "article_uuid" VARCHAR(36) UNIQUE PRIMARY KEY NOT NULL REFERENCES "article" ("uuid"),
        "title"        VARCHAR(256),
        "topic"        VARCHAR(32) REFERENCES "topic" ("id"),
        "slug"         VARCHAR(512),
        "read_time"    INT DEFAULT 0,
        "content"      TEXT
      );`,
    },
    {
      name: "article_link",
      definition: `
      CREATE TABLE "article_link"
      (
        "article_uuid"  VARCHAR(36) UNIQUE PRIMARY KEY NOT NULL REFERENCES "article" ("uuid"),
        "sharable_link" VARCHAR(248),
        "expires_at"    TIMESTAMP NOT NULL DEFAULT (datetime(current_timestamp, '+7 day'))
      );`,
    },
    {
      name: "tag",
      definition: `
      CREATE TABLE "tag"
      (
        "id"         VARCHAR(32) NOT NULL PRIMARY KEY,
        "name"       VARCHAR(32) NOT NULL,
        "created_at" TIMESTAMP NOT NULL DEFAULT current_timestamp,
        "updated_at" TIMESTAMP NOT NULL DEFAULT current_timestamp
      );`,
    },
    {
      name: "article_tag",
      definition: `
      CREATE TABLE "article_tag"
      (
        "article_uuid" VARCHAR(36) NOT NULL REFERENCES "article" ("uuid"),
        "tag_id"     VARCHAR(32) NOT NULL REFERENCES "tag" ("id")
      );`,
    },
  }

  var ctx = context.Background()
  tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    log.Fatal(err)
  }

  for _, t := range tables {
    if !t.exists(ctx, tx) {
      t.create(ctx, tx)
    }
  }

  if err = tx.Commit(); nil != err {
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
  engine.POST("/me.authenticate", me.Authenticate)
  engine.POST("/me.deauthenticate", me.Deauthenticate)

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

  engine.NoRoute(func(c *gin.Context) {
    var p problem.Problem
    p.Status(http.StatusNotFound)
    p.Title("Target not found.")
    p.Detail("Could not find the requested target resource. Possible causes: invalid URL, this resource no longer exists, or a temporary server issue.")
    p.Instance(c.Request.URL.String())
    p.Emit(c.Writer)
  })

  engine.HandleMethodNotAllowed = true
  var routes = engine.Routes()
  engine.NoMethod(func(c *gin.Context) {
    var allowedMethods = make([]string, 0, 1)
    for _, route := range routes {
      if route.Path == c.Request.URL.Path {
        allowedMethods = append(allowedMethods, route.Method)
      }
    }
    c.Header("Allow", strings.Join(allowedMethods, ","))
    var p problem.Problem
    p.Status(http.StatusMethodNotAllowed)
    p.Title("Unsupported HTTP method.")
    p.Detail(fmt.Sprintf("The target resource doesn't support this method (%s). Check the 'Allow' header in the response for a list of supported methods.", c.Request.Method))
    p.Instance(c.Request.URL.String())
    p.Emit(c.Writer)
  })

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
