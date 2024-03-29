package gormmw

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/pkg/errors"

	"github.com/gobuffalo/buffalo"
)

// Transaction is a piece of Buffalo middleware that wraps each
// request in a transaction. The transaction will automatically get
// committed if there's no errors and the response status code is a
// 2xx or 3xx, otherwise it'll be rolled back. It will also add a
// field to the log, "db", that shows the total duration spent during
// the request making database calls.
func Transaction(db *gorm.DB, logger gorm.LogWriter) buffalo.MiddlewareFunc {
	return func(h buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			// wrap all requests in a transaction and set the length
			// of time doing things in the db to the log.
			// ANY error returned by the tx function will cause the
			// tx to be rolled back
			couldBeDBorYourErr := func(tx *gorm.DB) error {
				// setup logging
				start := time.Now().Unix()
				defer func() {
					finished := time.Now().Unix()
					elapsed := time.Duration(finished - start)
					c.LogField("db", elapsed)
				}()

				// add the transaction to the context
				c.Set("tx", tx)

				// call the next handler; if it errors stop and return the error
				if yourError := h(c); yourError != nil {
					tx.Rollback()
					return yourError
				}

				// check the response status code. if the code is NOT 200..399
				// then it is considered "NOT SUCCESSFUL" and an error will be returned
				if res, ok := c.Response().(*buffalo.Response); ok {
					if res.Status < 200 || res.Status >= 400 {
						tx.Rollback()
						return errNonSuccess
					}
				}

				// as far was we can tell everything went well
				tx.Commit()
				return nil
			}(db.Begin())

			// couldBeDBorYourErr could be one of possible values:
			// * nil - everything went well, if so, return
			// * yourError - an error returned from your application, middleware, etc...
			// * a database error - this is returned if there were problems committing the transaction
			// * a errNonSuccess - this is returned if the response status code is not between 200..399
			if couldBeDBorYourErr != nil && errors.Cause(couldBeDBorYourErr) != errNonSuccess {
				return couldBeDBorYourErr
			}
			return nil
		}
	}
}

var errNonSuccess = errors.New("non success status code")
