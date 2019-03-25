package gormmw_test

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/httptest"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/yanshiyason/gormmw"
)

type widget struct {
	ID        uint       `gorm:"PRIMARY_KEY"`
	CreatedAt *time.Time `gorm:"not null"`
	UpdatedAt *time.Time `gorm:"not null"`
}

func tx(fn func(tx *gorm.DB)) error {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		return errors.WithStack(err)
	}
	path := filepath.Join(d, "pt_test.sqlite")
	defer os.RemoveAll(path)

	db, err := gorm.Open("sqlite3", path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Close()

	db.Debug()
	// db.LogMode(true)
	db.SetLogger(log.New(os.Stdout, "\n", 0))

	db.CreateTable(&widget{})

	fn(db)
	return nil
}

func app(db *gorm.DB) *buffalo.App {
	app := buffalo.New(buffalo.Options{})
	app.Use(gormmw.Transaction(db, log.New(os.Stdout, "\r\n", 0)))
	app.GET("/success", func(c buffalo.Context) error {
		w := &widget{}
		tx := c.Value("tx").(*gorm.DB)
		if db := tx.Create(w); db.Error != nil {
			return db.Error
		}
		return c.Render(201, nil)
	})
	app.GET("/non-success", func(c buffalo.Context) error {
		w := &widget{}
		tx := c.Value("tx").(*gorm.DB)
		if db := tx.Create(w); db.Error != nil {
			return db.Error
		}
		return c.Render(301, nil)
	})
	app.GET("/error", func(c buffalo.Context) error {
		w := &widget{}
		tx := c.Value("tx").(*gorm.DB)
		if db := tx.Create(w); db.Error != nil {
			return db.Error
		}
		return errors.New("boom")
	})
	return app
}

func Test_PopTransaction(t *testing.T) {
	r := require.New(t)
	err := tx(func(db *gorm.DB) {
		w := httptest.New(app(db))
		res := w.HTML("/success").Get()
		r.Equal(201, res.Code)
		var count int
		db.Table("widgets").Count(&count)
		r.Equal(1, count)
	})
	r.NoError(err)
}

func Test_PopTransaction_Error(t *testing.T) {
	r := require.New(t)
	err := tx(func(db *gorm.DB) {
		w := httptest.New(app(db))
		res := w.HTML("/error").Get()
		r.Equal(500, res.Code)
		var count int
		db.Table("widgets").Count(&count)
		r.Equal(0, count)
	})
	r.NoError(err)
}

func Test_PopTransaction_NonSuccess(t *testing.T) {
	r := require.New(t)
	err := tx(func(db *gorm.DB) {
		w := httptest.New(app(db))
		res := w.HTML("/non-success").Get()
		r.Equal(301, res.Code)
		var count int
		db.Table("widgets").Count(&count)
		r.Equal(1, count)
	})
	r.NoError(err)
}
