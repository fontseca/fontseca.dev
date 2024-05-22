package main

import (
  "context"
  "database/sql"
  "encoding/json"
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
  "path/filepath"
  "reflect"
  "strconv"
  "strings"
  "time"
)

// indentedWriter is a decorator that writes line-delimited JSON objects received from calls to
// the methods of the default slog.Logger to an io.Write, typically a *os.File; these
// writes are indented and the order of the object fields is conveniently rearranged.
type indentedWriter struct {
  io.Writer
}

func withIndentedWrite(w io.Writer) io.Writer {
  return &indentedWriter{Writer: w}
}

func (a *indentedWriter) Write(p []byte) (n int, err error) {
  n = len(p)

  if err = a.indent(&p); nil != err {
    return 0, err
  }

  _, err = a.Writer.Write(p) // discard n to avoid an 'io.ErrShortWrite' error in multiWriter.Write
  if nil != err {
    return 0, err
  }

  _, err = a.Writer.Write([]byte("\n"))
  if nil != err {
    return 0, err
  }

  return n, nil
}

func (a *indentedWriter) indent(p *[]byte) (err error) {
  var s = struct {
    Level  string `json:"level"`
    Time   string `json:"time"`
    Msg    string `json:"msg"`
    Source any    `json:"source"`
  }{}
  if err = json.Unmarshal(*p, &s); nil != err {
    return err
  }
  if *p, err = json.MarshalIndent(s, "", "  "); nil != err {
    return err
  }
  return nil
}

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
    err := db.Close()
    if err != nil {
      log.Fatal(err)
    }
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
      name: "article",
      definition: `
      CREATE TABLE "article"
      (
        "uuid"         VARCHAR(36) NOT NULL PRIMARY KEY DEFAULT (uuid_generate_v4 ()),
        "title"        VARCHAR(256) NOT NULL,
        "author"       VARCHAR(64) NOT NULL REFERENCES "me" ("username"),
        "slug"         VARCHAR(512) NOT NULL,
        "read_time"    INT NOT NULL ON CONFLICT REPLACE DEFAULT 0,
        "content"      TEXT NOT NULL ON CONFLICT REPLACE DEFAULT 'No content.',
        "draft"        BOOLEAN DEFAULT TRUE,
        "pinned"       BOOLEAN DEFAULT FALSE,
        "hidden"       BOOLEAN DEFAULT FALSE,
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
      name: "topic",
      definition: `
      CREATE TABLE "topic"
      (
        "uuid"       VARCHAR(36) NOT NULL PRIMARY KEY DEFAULT (uuid_generate_v4 ()),
        "name"       VARCHAR(64) NOT NULL,
        "created_at" TIMESTAMP NOT NULL DEFAULT current_timestamp,
        "updated_at" TIMESTAMP NOT NULL DEFAULT current_timestamp
      );`,
    },
    {
      name: "article_topic",
      definition: `
      CREATE TABLE "article_topic"
      (
        "article_uuid" VARCHAR(36) NOT NULL REFERENCES "article" ("uuid"),
        "topic_uuid"   VARCHAR(36) NOT NULL REFERENCES "topic" ("uuid")
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

  failLogFile, err := os.OpenFile("fail.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
  if nil != err {
    log.Fatal(err)
  }
  defer failLogFile.Close()

  var multiWriter = io.MultiWriter(withIndentedWrite(os.Stderr), failLogFile)
  var logger = slog.New(slog.NewJSONHandler(multiWriter,
    &slog.HandlerOptions{
      AddSource: true,
      ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
        if slog.SourceKey == a.Key {
          var source, _ = a.Value.Any().(*slog.Source)
          if nil != source {
            source.File = filepath.Base(source.File)
          }
        }
        return a
      },
    }))

  slog.SetDefault(logger)

  var mode = strings.TrimSpace(os.Getenv("SERVER_MODE"))
  if "" == mode {
    mode = gin.DebugMode
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
    log.Fatal(err)
  }
  defer serverLogFile.Close()

  engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
    Formatter: formatter,
    Output:    io.MultiWriter(os.Stdout, serverLogFile),
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
  engine.POST("/archive.drafts.topics.add", drafts.AddTopic)
  engine.POST("/archive.drafts.topics.remove", drafts.RemoveTopic)

  var (
    articlesService = service.NewArticlesService(archive)
    articles        = handler.NewArticlesHandler(articlesService)
  )

  engine.GET("archive.articles.list", articles.Get)
  engine.GET("archive.articles.hidden.list", articles.GetHidden)
  engine.GET("archive.articles.info", articles.GetByID)
  engine.POST("archive.articles.amend", articles.Amend)
  engine.POST("archive.articles.hide", articles.Hide)
  engine.POST("archive.articles.show", articles.Show)
  engine.POST("archive.articles.remove", articles.Remove)
  engine.POST("archive.articles.pin", articles.Pin)
  engine.POST("archive.articles.unpin", articles.Unpin)
  engine.POST("archive.articles.topics.add", articles.AddTopic)
  engine.POST("archive.articles.topics.remove", articles.RemoveTopic)

  var web = handler.NewWebHandler(
    meService,
    experienceService,
    projectsService,
  )

  engine.GET("/", web.RenderMe)
  engine.GET("/experience", web.RenderExperience)
  engine.GET("/work", web.RenderProjects)
  engine.GET("/work/:project_slug", web.RenderProjectDetails)

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

  var port = strings.TrimSpace(os.Getenv("SERVER_PORT"))
  if "" == port {
    port = ":5487"
  }

  var server = http.Server{
    Addr:           port,
    IdleTimeout:    1 * time.Minute,
    ReadTimeout:    5 * time.Second,
    WriteTimeout:   5 * time.Second,
    MaxHeaderBytes: 1024,
    Handler:        engine,
  }

  slog.Error(server.ListenAndServe().Error())
}
