package repository

import (
  "context"
  "database/sql"
  "fontseca.dev/model"
  "fontseca.dev/transfer"
  "github.com/google/uuid"
  "log/slog"
)

// ArticlesRepository is a common API for articles, article drafts
// and article patches.
//
// An article is a piece of writing about a particular subject in my
// website's archive. Naturally, every article has one or more topics
// that are inherent to the discussion of the article.
//
// An article draft, or just draft, is a rough version of an article
// that is not yet in its final form nor is it published. The main
// purpose of a draft is to provide a description of the main facts
// or points involved in the subject. This draft is improved by
// making revisions. You can share a draft in order to get feedback.
// Once the draft seems enticing and complete, it is published.
//
// When an article needs to be improved or amended, an article patch
// is created internally to record every enhancement made to the main
// article. The article is still available online during this process.
// An article patch, or simply patch, is an internal entity that points
// to the article it's been improving; it is a temporal place to store
// updates and enhancements. The patch is also shared to get feedback
// and improved by revisions. Once the patch is coherent and complete,
// it is released and physically merged to the original article.
//
// The draft and the article are both referenced by the same UUID. The
// patch is a completely different object that points to an article.
// Since an article can only have one patch at a time, by using the
// article's UUID, you can access any patch it currently has.
type ArticlesRepository interface {
  // Draft starts the creation process of an article. It returns the
  // UUID of the draft that was created.
  //
  // To draft an article, only its title is required, other fields
  // are completely optional and can be added in an eventual revision.
  Draft(ctx context.Context, creation *transfer.ArticleCreation) (id string, err error)

  // Publish makes a draft publicly available.
  //
  // Invoking Publish on an already published article or a patch has
  // no effect.
  Publish(ctx context.Context, id string) error

  // Get retrieves all the articles that are either hidden or not. If
  // draftsOnly is true, then only retrieves all the ongoing drafts.
  //
  // If needle is a non-empty string, then Get behaves like a search
  // function over non-hidden articles, so it attempts to find and
  // amass every article whose title or content contains any of the
  // keywords (if more than one) in needle.
  Get(ctx context.Context, needle string, hidden, draftsOnly bool) (articles []*model.Article, err error)

  // GetByID retrieves one article (or article draft) by its UUID.
  GetByID(ctx context.Context, id string, isDraft bool) (article *model.Article, err error)

  // Amend starts the process to update an article. To amend the article,
  // a public copy of it is kept available to everyone while a patch
  // is created to store any revision made to the article.
  //
  // If the article is still a draft, or it's already being amended,
  // any call to this method has no effect.
  Amend(ctx context.Context, id string) error

  // Remove completely removes an article and any patch it currently
  // has from the database. If the article is a draft, calling Remove
  // has no effect on it whatsoever.
  //
  // If you want to remove a draft, use Discard instead.
  Remove(ctx context.Context, id string) error

  // AddTopic adds a topic to the article. If the topic already
  // exists, it returns an error informing about a conflicting
  // state.
  AddTopic(ctx context.Context, articleID, topicID string) error

  // RemoveTopic removes a topic from the article. If the article has
  // no topic identified by its UUID, it returns an error indication
  // a not found state.
  RemoveTopic(ctx context.Context, articleID, topicID string) error

  // SetHidden hides or shows an article depending on the value of hidden.
  SetHidden(ctx context.Context, id string, hidden bool) error

  // SetPinned pins or unpins an article depending on the value of pinned.
  SetPinned(ctx context.Context, id string, pinned bool) error

  // Share creates a shareable link for a draft or a patch. Only users
  // with that link can see the progress and provide feedback.
  //
  // A shareable link does not make an article public. This link will
  // eventually expire after a certain amount of time.
  Share(ctx context.Context, id string) (link string, err error)

  // Discard completely drops a draft; otherwise if called on a patch
  // it discards it but keeps the original article.
  //
  // This method has no effect on an article.
  Discard(ctx context.Context, id string) error

  // Revise adds a correction or inclusion to a draft or patch in order
  // to correct or improve it.
  Revise(ctx context.Context, id string) error

  // Release merges patch into the original article and published the
  // update immediately after merging.
  //
  // This method works only for patches.
  Release(ctx context.Context, id string) error

  // GetPatches retrieves all the ongoing patches of every article.
  GetPatches(ctx context.Context) (patches []any, err error)
}

type articlesRepository struct {
  db *sql.DB
}

func NewArticlesRepository(db *sql.DB) ArticlesRepository {
  return &articlesRepository{db}
}

func (r *articlesRepository) Draft(ctx context.Context, creation *transfer.ArticleCreation) (id string, err error) {
  tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return uuid.Nil.String(), err
  }

  defer tx.Rollback()

  draftArticleQuery := `
  INSERT INTO "article" ("title", "author", "slug", "read_time", "content")
                 VALUES (@title, 'fontseca.dev', @slug, @read_time, @content)
              RETURNING "uuid";`

  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  result := tx.QueryRowContext(ctx, draftArticleQuery,
    sql.Named("title", creation.Title),
    sql.Named("slug", creation.Slug),
    sql.Named("read_time", creation.ReadTime),
    sql.Named("content", creation.Content),
  )

  if err = result.Scan(&id); nil != err {
    slog.Error(err.Error())
    return uuid.Nil.String(), err
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return uuid.Nil.String(), err
  }

  return id, nil
}

func (r *articlesRepository) Publish(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) Get(ctx context.Context, needle string, hidden, draftsOnly bool) (articles []*model.Article, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) GetByID(ctx context.Context, id string, isDraft bool) (article *model.Article, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) Amend(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) Remove(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) AddTopic(ctx context.Context, articleID, topicID string) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) RemoveTopic(ctx context.Context, articleID, topicID string) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) SetHidden(ctx context.Context, id string, hidden bool) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) SetPinned(ctx context.Context, id string, pinned bool) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) Share(ctx context.Context, id string) (link string, err error) {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) Discard(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) Revise(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) Release(ctx context.Context, id string) error {
  // TODO implement me
  panic("implement me")
}

func (r *articlesRepository) GetPatches(ctx context.Context) (patches []any, err error) {
  // TODO implement me
  panic("implement me")
}
