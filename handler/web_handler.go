package handler

import (
  "context"
  "database/sql"
  "errors"
  "fontseca.dev/components/pages"
  "fontseca.dev/components/ui"
  "fontseca.dev/model"
  "fontseca.dev/repository"
  "fontseca.dev/service"
  "fontseca.dev/transfer"
  "github.com/gin-gonic/gin"
  "net/http"
  "slices"
  "strconv"
  "strings"
  "time"
)

type WebHandler struct {
  meService         service.MeService
  experienceService service.ExperienceService
  projectsService   service.ProjectsService
  drafts            service.DraftsService
  articles          service.ArticlesService
  topics            service.TopicsService
  tags              service.TagsService
}

func NewWebHandler(
  meService service.MeService,
  experienceService service.ExperienceService,
  projectsService service.ProjectsService,
  drafts service.DraftsService,
  articles service.ArticlesService,
  topics service.TopicsService,
  tags service.TagsService,
) *WebHandler {
  return &WebHandler{
    meService:         meService,
    experienceService: experienceService,
    projectsService:   projectsService,
    drafts:            drafts,
    articles:          articles,
    topics:            topics,
    tags:              tags,
  }
}

func (h *WebHandler) internal(c *gin.Context) {
  c.Status(http.StatusInternalServerError)
  http.Error(c.Writer, "500 Internal Server Error", http.StatusInternalServerError)
}

func (h *WebHandler) RenderMe(c *gin.Context) {
  me, err := h.meService.Get(c)
  if nil != err {
    return
  }
  pages.Me(me).Render(c, c.Writer)
}

func (h *WebHandler) RenderExperience(c *gin.Context) {
  var exp, err = h.experienceService.Get(c)
  if nil != err {
    return
  }
  pages.Experience(exp).Render(c, c.Writer)
}

func (h *WebHandler) RenderProjects(c *gin.Context) {
  var projects, err = h.projectsService.Get(c, false)
  if nil != err {
    return
  }
  pages.Projects(projects).Render(c, c.Writer)
}

func (h *WebHandler) RenderProjectDetails(c *gin.Context) {
  slug := c.Param("project_slug")
  var project, err = h.projectsService.GetBySlug(c, slug)
  if nil != err {
    c.Status(http.StatusNotFound)
    pages.ProjectDetails(nil).Render(c, c.Writer)
    return
  }
  pages.ProjectDetails(project).Render(c, c.Writer)
}

func (h *WebHandler) RenderArchive(c *gin.Context) {
  var (
    anyTopicSentinel      = &model.Topic{ID: "any", Name: "Any topic"}
    search, includeSearch = c.GetQuery("search")
    year, _               = strconv.Atoi(c.Param("year"))
    month, _              = strconv.Atoi(c.Param("month"))
    topic, includeTopic   = c.Params.Get("topic")
    filter                = &transfer.ArticleFilter{
      Search:      strings.TrimSpace(search),
      Topic:       topic,
      Publication: &transfer.Publication{Month: time.Month(month), Year: year},
      Page:        1,
      RPP:         10000,
    }
  )

  if anyTopicSentinel.ID == topic {
    filter.Topic = ""
  }

  articles, err := h.articles.Get(c, filter)

  if nil != err {
    h.internal(c)
    return
  }

  publications, err := h.articles.Publications(c)

  if nil != err {
    h.internal(c)
    return
  }

  topics, err := h.topics.Get(c)

  if nil != err {
    h.internal(c)
    return
  }

  hxRequest, _ := strconv.ParseBool(c.GetHeader("HX-Request"))

  if hxRequest && (includeSearch || includeTopic) {
    ui.SearchResults(articles).Render(c, c.Writer)
    return
  }

  var (
    i             = slices.IndexFunc(topics, func(t *model.Topic) bool { return t.ID == topic })
    selectedTopic = anyTopicSentinel
  )

  if -1 != i {
    selectedTopic = topics[i]
  } else {
    if "" != topic {
      selectedTopic.Name = "?"
    }
  }

  tags, err := h.tags.Get(c)

  if nil != err {
    h.internal(c)
    return
  }

  pages.Archive(
    articles,
    publications,
    topics,
    tags,
    filter.Search,
    filter.Publication,
    selectedTopic,
  ).Render(c, c.Writer)
}

func (h *WebHandler) RenderArticle(c *gin.Context) {
  cc := context.WithValue(c.Request.Context(), repository.VisitorKey, c.RemoteIP())
  c.Request = c.Request.Clone(cc)

  if _, checksum := c.Params.Get("hash"); checksum {
    shareableLink := c.Request.URL.Path

    if '/' != shareableLink[0] {
      shareableLink = "/" + shareableLink
    }

    draft, err := h.drafts.GetByLink(c.Request.Context(), shareableLink)

    if nil != err {
      switch {
      default:
        http.Error(c.Writer, "500 Internal Server Error", http.StatusInternalServerError)
        return
      case strings.Contains(err.Error(), "has expired") ||
        strings.Contains(err.Error(), "might have been either removed or blocked."):
        http.Error(c.Writer, "404 Not Found", http.StatusNotFound)
        return
      }
    }

    pages.Article(draft).Render(c, c.Writer)
    return
  }

  topic := c.Param("topic")
  year, _ := strconv.Atoi(c.Param("year"))
  month, _ := strconv.Atoi(c.Param("month"))
  slug := c.Param("slug")

  r := &transfer.ArticleRequest{
    Topic: topic,
    Publication: &transfer.Publication{
      Month: time.Month(month),
      Year:  year,
    },
    Slug: slug,
  }

  article, err := h.articles.GetOne(c.Request.Context(), r)

  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      http.Error(c.Writer, "404 Not Found", http.StatusNotFound)
      return
    } else {
      h.internal(c)
    }
    return
  }

  pages.Article(article).Render(c, c.Writer)
}
