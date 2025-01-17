package main

import (
	context "context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type dbmap struct {
	dbfs []*dbfld
	tbln string
}
type dbfld struct {
	fld string
	qry string
	col string
	typ string
	dfl string
	nul bool
	upd bool
}

func new_dbmap[T any]() *dbmap {
	dbm := &dbmap{dbfs: []*dbfld{}}
	obj := new(T)
	rfl := &rflt{}
	flds := rfl.fields(obj)
	for _, fld := range flds {
		dbm.dbfs = append(dbm.dbfs, &dbfld{fld: fld})
	}
	return dbm
}

// column() will define a custom mapping for fld. This function must be called before Table().
func (dbm *dbmap) column(fld, col, qry string) {
	for _, dbf := range dbm.dbfs {
		if strings.EqualFold(dbf.fld, fld) {
			// We found the dbfld by using the fld passed in.
			dbf.col = col
			dbf.qry = qry
			return
		}
	}
}

func (dbm *dbmap) mapped() []*dbfld {
	flds := []*dbfld{}
	for _, fld := range dbm.dbfs {
		if fld.col != "" && fld.fld != "" {
			flds = append(flds, fld)
		}
	}
	return flds
}
func (dbm *dbmap) mapped_upd() []*dbfld {
	flds := []*dbfld{}
	for _, fld := range dbm.dbfs {
		if fld.col != "" && fld.fld != "" && fld.upd {
			flds = append(flds, fld)
		}
	}
	return flds
}
func (dbm *dbmap) table(pool *pgxpool.Pool, tbln string) {
	// What are the actual column names on this table?
	dbm.tbln = tbln
	schm := ""
	_tbl := tbln
	toks := strings.Split(tbln, ".")
	if len(toks) > 1 {
		schm = toks[0]
		_tbl = toks[1]
	}
	qry := fmt.Sprintf("select column_name, data_type, is_nullable, column_default from information_schema.columns where table_schema = '%s' and table_name = '%s';", schm, _tbl)
	if rows, err := pool.Query(context.Background(), qry); err == nil {
		for rows.Next() {
			var coln, colt, null string
			var dflt sql.NullString
			rows.Scan(&coln, &colt, &null, &dflt)
			// Find the existing dbfld - it must exist (indicates the field on the object).
			for _, dbf := range dbm.dbfs {
				// Implicit match on fld (no custom mapping) or on col (explicit mapping).
				if strings.EqualFold(dbf.fld, coln) || strings.EqualFold(dbf.col, coln) {
					dbf.nul = strings.EqualFold(null, "YES")
					dbf.typ = colt
					dbf.upd = true
					dbf.col = coln		// In case not set, and we matched on dbf.fld
					// Fill in remaining db info.
					if dflt.Valid {
						if dflt.String == "''::text" {
							dbf.dfl = "''"
						} else if dflt.String == "'{}'::text[]" {
							dbf.dfl = "'{}'"
						} else {
							dbf.dfl = dflt.String
						}
						if strings.HasPrefix(dflt.String, "nextval") {
							dbf.upd = false
						}
					}
					break
				}
			}
		}
		rows.Close()
		// Now remove dbf entries (object fields) that have no corresponding database column.
		// dbfs := []*dbfld{}
		// for _, dbf := range dbm.dbfs {
		// 	if dbf.col != "" {
		// 		dbfs = append(dbfs, dbf)
		// 	}
		// }
		// dbm.dbfs = dbfs
	}
}

func (dbm *dbmap) find(coln string, mustExist, mustMapped bool) *dbfld {
	for _, dbf := range dbm.dbfs {
		if strings.EqualFold(dbf.col, coln) {
			if dbf.fld != "" {
				return dbf
			}
			if mustMapped {
				dbm.Print()
				panic(fmt.Sprintf("tbln[%s] coln[%s] not mapped", dbm.tbln, coln))
			}
			return nil
		}
	}
	if mustExist {
		dbm.Print()
		panic(fmt.Sprintf("tbln[%s] coln[%s] not found", dbm.tbln, coln))
	}
	return nil
}

func (dbm *dbmap) getColumnValueAsString(coln, fv string) string {
	for _, dbf := range dbm.dbfs {
		if strings.EqualFold(dbf.col, coln) {
			if fv == "" {
				if dbf.dfl != "" {
					return dbf.dfl
				} else if dbf.nul {
					return "NULL"
				}
			}
			if strings.HasPrefix(dbf.typ, "timestamp") {
				if fv == "" || fv == "NULL" || fv == "0" {
					if dbf.dfl != "" {
						return dbf.dfl
					} else if dbf.nul {
						return "NULL"
					}
					return ""
				}
				if i64, err := strconv.ParseInt(fv, 10, 64); err == nil {
					secs := i64 / (1000 * 1000)
					micr := i64 % (1000 * 1000)
					tm   := time.Unix(secs, micr * 1000)
					return fmt.Sprintf("'%s'", tm.Format("2006-01-02 15:04:05.000000"))
				} else {
					panic(fmt.Sprintf("col=%s fld=%s err=%s", dbf.col, dbf.fld, err.Error()))
				}
			}
			if strings.HasPrefix(dbf.typ, "time") {
				if fv == "" || fv == "NULL" || fv == "0" {
					if dbf.dfl != "" {
						return dbf.dfl
					} else if dbf.nul {
						return "NULL"
					}
					return ""
				}
				if i64, err := strconv.ParseInt(fv, 10, 64); err == nil {
					secs := i64 / (1000 * 1000)
					micr := i64 % (1000 * 1000)
					tm   := time.Unix(secs, micr * 1000)
					return fmt.Sprintf("'%s'", tm.Format("15:04:05.000000"))
				} else {
					panic(fmt.Sprintf("col=%s fld=%s err=%s", dbf.col, dbf.fld, err.Error()))
				}
			}
			if strings.HasPrefix(dbf.typ, "date") {
				if fv == "" || fv == "NULL" || fv == "0" {
					if dbf.dfl != "" {
						return dbf.dfl
					} else if dbf.nul {
						return "NULL"
					}
					return ""
				}
				if i64, err := strconv.ParseInt(fv, 10, 64); err == nil {
					secs := i64 / (1000 * 1000)
					micr := i64 % (1000 * 1000)
					tm   := time.Unix(secs, micr * 1000)
					return fmt.Sprintf("'%s'", tm.Format("2006-01-02"))
				} else {
					panic(fmt.Sprintf("col=%s fld=%s err=%s", dbf.col, dbf.fld, err.Error()))
				}
			}
			if strings.EqualFold(dbf.typ, "ARRAY") {
				// Special handling if the column type is an array
				// The fv will already be comma-separated, so just need to wrap
				// Weirdness - seems sometimes there can be blank entries, so remove them first.
				toks := strings.Split(fv, ",")
				list := []string{}
				for _, tok := range toks {
					if tok != "" {
						list = append(list, tok)
					}
				}
				return fmt.Sprintf("'{%s}'", strings.Join(list, ","))
			}
		}
	}
	return fmt.Sprintf("'%s'", fv)
}

func (dbm *dbmap) Print() {
	fmt.Printf("tbln=%s\n", dbm.tbln)
	fmt.Printf("%-6s %-25s %-40s %-6s %-6s %-10s %s\n", "fld", "col", "typ", "upd", "nul", "dfl", "qry")
	for _, dbf := range dbm.dbfs {
		fmt.Printf("%-6s %-25s %-40s %-6s %-6s %-10s %s\n", dbf.fld, dbf.col, dbf.typ, strconv.FormatBool(dbf.upd), strconv.FormatBool(dbf.nul), dbf.dfl, dbf.qry)
	}
}
