This is middleware was adapted from popmw to work with gorm. Original source code:
https://github.com/gobuffalo/buffalo-pop/tree/master/pop/popmw


## Usage:

The only difference in usage, is that the logger must be passed in as an argument.
In your buffalo app:

```golang
// Wraps each request in a transaction.
//  c.Value("tx").(*gorm.DB)
// Remove to disable this.
app.Use(gormmw.Transaction(models.DB, app.Logger))
```
