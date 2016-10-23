package controllers

import (
  "database/sql"
  "github.com/go-gorp/gorp"
  _ "github.com/lib/pq"
  r "github.com/revel/revel"
  "github.com/revel/modules/db/app"
  "credREST/app/models"
)

var (
  Dbm *gorp.DbMap
)

func InitDB() {
  db.Init()
  Dbm = &gorp.DbMap{Db: db.Db, Dialect: gorp.PostgresDialect{}}

  setColumnSizes := func(t *gorp.TableMap, colSizes map[string]int) {
    for col, size := range colSizes {
      t.ColMap(col).MaxSize = size
    }
  }

  t := Dbm.AddTable(models.University{}).SetKeys(true, "UniversityId")
  t.ColMap("Degrees").Transient = true
  setColumnSizes(t, map[string]int{
    "Name":     50,
    "Country":  50,
  })

  t = Dbm.AddTable(models.Degree{}).SetKeys(true, "DegreeId")
  t.ColMap("University").Transient = true
  t.ColMap("Courses").Transient = true
  setColumnSizes(t, map[string]int{
    "Name": 50,
    "TotalCredits": 20,
  })

  t = Dbm.AddTable(models.Course{}).SetKeys(true, "CourseId")
  t.ColMap("Degree").Transient = true
  setColumnSizes(t, map[string]int{
    "Name":     50,
    "Credits":  20,
  })

  Dbm.TraceOn("[gorp]", r.INFO)
  Dbm.CreateTables()

  university := &models.University{0, "UdelaR", "Uruguay", nil}
  if err := Dbm.Insert(university); err != nil {
    panic(err)
  }

  degree := &models.Degree{0, 0, "Ingenieria en Computacion", 450, university, nil}
  if err := Dbm.Insert(degree); err != nil {
    panic(err)
  }

  course := &models.Course{0, 0, "MAA", 10, degree}
  if err := Dbm.Insert(course); err != nil {
    panic(err)
  }

}

type GorpController struct {
  *r.Controller
  Txn *gorp.Transaction
}

func (c *GorpController) Begin() r.Result {
  txn, err := Dbm.Begin()
  if err != nil {
    panic(err)
  }
  c.Txn = txn
  return nil
}

func (c *GorpController) Commit() r.Result {
  if c.Txn == nil {
    return nil
  }
  if err := c.Txn.Commit(); err != nil && err != sql.ErrTxDone {
    panic(err)
  }
  c.Txn = nil
  return nil
}

func (c *GorpController) Rollback() r.Result {
  if c.Txn == nil {
    return nil
  }
  if err := c.Txn.Rollback(); err != nil && err != sql.ErrTxDone {
    panic(err)
  }
  c.Txn = nil
  return nil
}
