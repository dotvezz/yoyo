package template

const (
	QueryPackageName      = "$ENTITY_PACKAGE_NAME$"
	FieldName             = "$PK_FIELD_NAME$"
	PKFields              = "$PK_FIELDS"
	Type                  = "$TYPE$"
)

const NoPKCapture = `
	_ = res
`

var SinglePKCaptureTemplate = `
	e = in
	var eid int64
	eid, err = res.LastInsertId()
	e.` + FieldName + ` = ` + Type + `(eid)
	if err != nil {
		return e, err
	}
`

const MultiPKCaptureTemplate = `
	e = in
	var eid int64
	eid, err = res.LastInsertId()
	e.Id = int32(eid)
	if err != nil {
		return e, err
	}
`

const PKQueryTemplate = `
	q, args := ` + QueryPackageName + `.Query{}.
		` + PKFields + `
		SQL()
`

const PKFieldTemplate = FieldName + "(in.persisted." + FieldName + ")."
