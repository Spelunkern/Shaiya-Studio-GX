package sdata

import (
	"fmt"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
)

// QuestReq holds the requirements and objectives of a quest (QuestReq).
type QuestReq struct {
	MinLevel, MaxLevel      int16
	Country, Grow, Na2, Na3 uint8
	Classes                 [6]uint8
	Prv2, Prv3, Na4         int16
	Na5                     uint8
	Prv1                    int16
	Na6, U1C                int32
	U20                     int16
	U22                     uint8
	Timer                   int16
	U28, U2C                int32
	U30                     int16
	Rept, U33               uint8
	U34                     int16
	Qmod, Nty1              uint8
	Nti1                    int16
	Ity0, Iti0              uint8
	Rty1, Rti1, Rcat1       uint8
	Rty2, Rti2, Rcat2       uint8
	Rty3, Rti3, Rcat3       uint8
	Qtyp, Nty2              uint8
	Nti2                    int16
	Fty1, Fti1, Fct1        uint8
	Fty2, Fti2, Fct2        uint8
	Fty3, Fti3, Fct3        uint8
	Kpvp                    uint8
	Mbi1                    int16
	Mct1                    uint8
	Mbi2                    int16
	Mct2, Drwd, Clrw        uint8
}

// fields lists the fields in file order.
func (q *QuestReq) fields() []any {
	f := []any{&q.MinLevel, &q.MaxLevel, &q.Country, &q.Grow, &q.Na2, &q.Na3}
	for i := range q.Classes {
		f = append(f, &q.Classes[i])
	}
	return append(f, &q.Prv2, &q.Prv3, &q.Na4, &q.Na5, &q.Prv1, &q.Na6, &q.U1C, &q.U20, &q.U22, &q.Timer,
		&q.U28, &q.U2C, &q.U30, &q.Rept, &q.U33, &q.U34, &q.Qmod, &q.Nty1, &q.Nti1, &q.Ity0, &q.Iti0,
		&q.Rty1, &q.Rti1, &q.Rcat1, &q.Rty2, &q.Rti2, &q.Rcat2, &q.Rty3, &q.Rti3, &q.Rcat3, &q.Qtyp, &q.Nty2,
		&q.Nti2, &q.Fty1, &q.Fti1, &q.Fct1, &q.Fty2, &q.Fti2, &q.Fct2, &q.Fty3, &q.Fti3, &q.Fct3, &q.Kpvp,
		&q.Mbi1, &q.Mct1, &q.Mbi2, &q.Mct2, &q.Drwd, &q.Clrw)
}

func readField(r *binstream.Reader, p any) {
	switch v := p.(type) {
	case *uint8:
		*v = r.Byte()
	case *int16:
		*v = r.Int16()
	case *uint16:
		*v = r.Uint16()
	case *int32:
		*v = r.Int32()
	}
}

func writeField(w *binstream.Writer, p any) {
	switch v := p.(type) {
	case *uint8:
		w.Byte(*v)
	case *int16:
		w.Int16(*v)
	case *uint16:
		w.Uint16(*v)
	case *int32:
		w.Int32(*v)
	}
}

func (q *QuestReq) read(r *binstream.Reader) {
	for _, p := range q.fields() {
		readField(r, p)
	}
}

func (q *QuestReq) write(w *binstream.Writer) {
	for _, p := range q.fields() {
		writeField(w, p)
	}
}

// text exports the requirements, each value preceded by sep; the reward
// and objective triples are tab separated (QuestReq::toText).
func (q *QuestReq) text(sep string) string {
	var b strings.Builder
	v := func(x ...any) {
		for _, e := range x {
			b.WriteString(sep)
			fmt.Fprint(&b, e)
		}
	}
	triple := func(a, c, d uint8) {
		b.WriteString(sep)
		fmt.Fprintf(&b, "%d\t%d\t%d", a, c, d)
	}
	v(q.MinLevel, q.MaxLevel, q.Country, q.Grow, q.Na2, q.Na3)
	for _, c := range q.Classes {
		v(c)
	}
	v(q.Prv2, q.Prv3, q.Na4, q.Na5, q.Prv1, q.Na6, q.U1C, q.U20, q.U22, q.Timer, q.U28, q.U2C, q.U30, q.Rept,
		q.U33, q.U34, q.Qmod, q.Nty1, q.Nti1, q.Ity0, q.Iti0)
	triple(q.Rty1, q.Rti1, q.Rcat1)
	triple(q.Rty2, q.Rti2, q.Rcat2)
	triple(q.Rty3, q.Rti3, q.Rcat3)
	v(q.Qtyp, q.Nty2, q.Nti2)
	triple(q.Fty1, q.Fti1, q.Fct1)
	triple(q.Fty2, q.Fti2, q.Fct2)
	triple(q.Fty3, q.Fti3, q.Fct3)
	v(q.Kpvp, q.Mbi1, q.Mct1, q.Mbi2, q.Mct2, q.Drwd, q.Clrw)
	b.WriteString("\r\n")
	return b.String()
}

// QuestResult is one outcome of a quest (QuestResult).
type QuestResult struct {
	Mk, Mu, Mc, Ik, Ic, Iu uint8
	U1, U2, Exp, Gold      int32
	T, I, C                [3]uint8
	Unl                    int16
}

func (q *QuestResult) read(r *binstream.Reader, version int) {
	q.Mk, q.Mu, q.Mc, q.Ik, q.Ic, q.Iu = r.Byte(), r.Byte(), r.Byte(), r.Byte(), r.Byte(), r.Byte()
	q.U1, q.U2, q.Exp, q.Gold = r.Int32(), r.Int32(), r.Int32(), r.Int32()
	if (q.Mu != 0 || q.Iu != 0) && q.Mk == 0 && q.Ik == 0 {
		*q = QuestResult{}
	}
	q.T[0] = r.Byte()
	q.I[0] = r.Byte()
	if version&0xFFF >= 0x600 {
		q.C[0] = r.Byte()
		q.T[1], q.I[1], q.C[1] = r.Byte(), r.Byte(), r.Byte()
		q.T[2], q.I[2], q.C[2] = r.Byte(), r.Byte(), r.Byte()
	}
	q.Unl = r.Int16()
}

func (q *QuestResult) write(w *binstream.Writer, version int) {
	w.Put([]byte{q.Mk, q.Mu, q.Mc, q.Ik, q.Ic, q.Iu})
	w.Int32(q.U1)
	w.Int32(q.U2)
	w.Int32(q.Exp)
	w.Int32(q.Gold)
	if version&0xFFF >= 0x600 {
		w.Put([]byte{q.T[0], q.I[0], q.C[0], q.T[1], q.I[1], q.C[1], q.T[2], q.I[2], q.C[2]})
	} else {
		w.Put([]byte{q.T[0], q.I[0]})
	}
	w.Int16(q.Unl)
}

// QuestDetails is a quest of NpcQuest.SData (QuestDetails).
type QuestDetails struct {
	Version    int
	QuestID    int16
	Name       string
	Summary    string
	Replies    [6]string
	Desc       string
	Welcome    string
	Reminder   string
	AltMessage string
	Req        QuestReq
	Results    [6]QuestResult
}

// NewQuest returns an empty quest.
func NewQuest(version int) *QuestDetails { return &QuestDetails{Version: version} }

// Reset clears the texts, requirements and results.
func (q *QuestDetails) Reset() {
	*q = QuestDetails{Version: q.Version, QuestID: q.QuestID}
}

func (q *QuestDetails) strings() []*string {
	return []*string{&q.Name, &q.Summary, &q.Desc, &q.Welcome, &q.Reminder, &q.AltMessage,
		&q.Replies[0], &q.Replies[1], &q.Replies[2], &q.Replies[3], &q.Replies[4], &q.Replies[5]}
}

// ltxtStrings lists the texts in the order of NpcQuestTrans files: name,
// summary, the six replies, then description, welcome, reminder and
// alternate message (the order QuestDetails::write uses).
func (q *QuestDetails) ltxtStrings() []*string {
	return []*string{&q.Name, &q.Summary, &q.Replies[0], &q.Replies[1], &q.Replies[2], &q.Replies[3],
		&q.Replies[4], &q.Replies[5], &q.Desc, &q.Welcome, &q.Reminder, &q.AltMessage}
}

func (q *QuestDetails) ptr(tag Tag) any {
	switch tag {
	case "quid":
		return &q.QuestID
	case "lmin":
		return &q.Req.MinLevel
	case "lmax":
		return &q.Req.MaxLevel
	case "cntr":
		return &q.Req.Country
	case "grow":
		return &q.Req.Grow
	case "na2 ":
		return &q.Req.Na2
	case "na3 ":
		return &q.Req.Na3
	case "cla1":
		return &q.Req.Classes[0]
	case "cla2":
		return &q.Req.Classes[1]
	case "cla3":
		return &q.Req.Classes[2]
	case "cla4":
		return &q.Req.Classes[3]
	case "cla5":
		return &q.Req.Classes[4]
	case "cla6":
		return &q.Req.Classes[5]
	case "prv1":
		return &q.Req.Prv1
	case "prv2":
		return &q.Req.Prv2
	case "prv3":
		return &q.Req.Prv3
	case "na4 ":
		return &q.Req.Na4
	case "na5 ":
		return &q.Req.Na5
	case "\x00na6":
		return &q.Req.Na6
	case "timr":
		return &q.Req.Timer
	case "rept":
		return &q.Req.Rept
	case "qmod":
		return &q.Req.Qmod
	case "nty1":
		return &q.Req.Nty1
	case "nti1":
		return &q.Req.Nti1
	case "ity0":
		return &q.Req.Ity0
	case "iti0":
		return &q.Req.Iti0
	case "qtyp":
		return &q.Req.Qtyp
	case "nty2":
		return &q.Req.Nty2
	case "nti2":
		return &q.Req.Nti2
	case "rty1":
		return &q.Req.Rty1
	case "rti1":
		return &q.Req.Rti1
	case "rct1":
		return &q.Req.Rcat1
	case "rty2":
		return &q.Req.Rty2
	case "rti2":
		return &q.Req.Rti2
	case "rct2":
		return &q.Req.Rcat2
	case "rty3":
		return &q.Req.Rty3
	case "rti3":
		return &q.Req.Rti3
	case "rct3":
		return &q.Req.Rcat3
	case "fty1":
		return &q.Req.Fty1
	case "fti1":
		return &q.Req.Fti1
	case "fct1":
		return &q.Req.Fct1
	case "fty2":
		return &q.Req.Fty2
	case "fti2":
		return &q.Req.Fti2
	case "fct2":
		return &q.Req.Fct2
	case "fty3":
		return &q.Req.Fty3
	case "fti3":
		return &q.Req.Fti3
	case "fct3":
		return &q.Req.Fct3
	case "kpvp":
		return &q.Req.Kpvp
	case "mbi1":
		return &q.Req.Mbi1
	case "mbi2":
		return &q.Req.Mbi2
	case "mct1":
		return &q.Req.Mct1
	case "mct2":
		return &q.Req.Mct2
	case "drwd":
		return &q.Req.Drwd
	case "clrw":
		return &q.Req.Clrw
	case "r1mk":
		return &q.Results[0].Mk
	case "r2mk":
		return &q.Results[1].Mk
	case "r3mk":
		return &q.Results[2].Mk
	case "r4mk":
		return &q.Results[3].Mk
	case "r5mk":
		return &q.Results[4].Mk
	case "r6mk":
		return &q.Results[5].Mk
	case "r1mu":
		return &q.Results[0].Mu
	case "r2mu":
		return &q.Results[1].Mu
	case "r3mu":
		return &q.Results[2].Mu
	case "r4mu":
		return &q.Results[3].Mu
	case "r5mu":
		return &q.Results[4].Mu
	case "r6mu":
		return &q.Results[5].Mu
	case "r1mc":
		return &q.Results[0].Mc
	case "r2mc":
		return &q.Results[1].Mc
	case "r3mc":
		return &q.Results[2].Mc
	case "r4mc":
		return &q.Results[3].Mc
	case "r5mc":
		return &q.Results[4].Mc
	case "r6mc":
		return &q.Results[5].Mc
	case "r1ik":
		return &q.Results[0].Ik
	case "r2ik":
		return &q.Results[1].Ik
	case "r3ik":
		return &q.Results[2].Ik
	case "r4ik":
		return &q.Results[3].Ik
	case "r5ik":
		return &q.Results[4].Ik
	case "r6ik":
		return &q.Results[5].Ik
	case "r1iu":
		return &q.Results[0].Iu
	case "r2iu":
		return &q.Results[1].Iu
	case "r3iu":
		return &q.Results[2].Iu
	case "r4iu":
		return &q.Results[3].Iu
	case "r5iu":
		return &q.Results[4].Iu
	case "r6iu":
		return &q.Results[5].Iu
	case "r1ic":
		return &q.Results[0].Ic
	case "r2ic":
		return &q.Results[1].Ic
	case "r3ic":
		return &q.Results[2].Ic
	case "r4ic":
		return &q.Results[3].Ic
	case "r5ic":
		return &q.Results[4].Ic
	case "r6ic":
		return &q.Results[5].Ic
	case "r1u1":
		return &q.Results[0].U1
	case "r2u1":
		return &q.Results[1].U1
	case "r3u1":
		return &q.Results[2].U1
	case "r4u1":
		return &q.Results[3].U1
	case "r5u1":
		return &q.Results[4].U1
	case "r6u1":
		return &q.Results[5].U1
	case "r1u2":
		return &q.Results[0].U2
	case "r2u2":
		return &q.Results[1].U2
	case "r3u2":
		return &q.Results[2].U2
	case "r4u2":
		return &q.Results[3].U2
	case "r5u2":
		return &q.Results[4].U2
	case "r6u2":
		return &q.Results[5].U2
	case "gld1":
		return &q.Results[0].Gold
	case "gld2":
		return &q.Results[1].Gold
	case "gld3":
		return &q.Results[2].Gold
	case "gld4":
		return &q.Results[3].Gold
	case "gld5":
		return &q.Results[4].Gold
	case "gld6":
		return &q.Results[5].Gold
	case "xp1 ":
		return &q.Results[0].Exp
	case "xp2 ":
		return &q.Results[1].Exp
	case "xp3 ":
		return &q.Results[2].Exp
	case "xp4 ":
		return &q.Results[3].Exp
	case "xp5 ":
		return &q.Results[4].Exp
	case "xp6 ":
		return &q.Results[5].Exp
	case "r1t1":
		return &q.Results[0].T[0]
	case "r1t2":
		return &q.Results[0].T[1]
	case "r1t3":
		return &q.Results[0].T[2]
	case "r2t1":
		return &q.Results[1].T[0]
	case "r2t2":
		return &q.Results[1].T[1]
	case "r2t3":
		return &q.Results[1].T[2]
	case "r3t1":
		return &q.Results[2].T[0]
	case "r3t2":
		return &q.Results[2].T[1]
	case "r3t3":
		return &q.Results[2].T[2]
	case "r4t1":
		return &q.Results[3].T[0]
	case "r4t2":
		return &q.Results[3].T[1]
	case "r4t3":
		return &q.Results[3].T[2]
	case "r5t1":
		return &q.Results[4].T[0]
	case "r5t2":
		return &q.Results[4].T[1]
	case "r5t3":
		return &q.Results[4].T[2]
	case "r6t1":
		return &q.Results[5].T[0]
	case "r6t2":
		return &q.Results[5].T[1]
	case "r6t3":
		return &q.Results[5].T[2]
	case "r1i1":
		return &q.Results[0].I[0]
	case "r1i2":
		return &q.Results[0].I[1]
	case "r1i3":
		return &q.Results[0].I[2]
	case "r2i1":
		return &q.Results[1].I[0]
	case "r2i2":
		return &q.Results[1].I[1]
	case "r2i3":
		return &q.Results[1].I[2]
	case "r3i1":
		return &q.Results[2].I[0]
	case "r3i2":
		return &q.Results[2].I[1]
	case "r3i3":
		return &q.Results[2].I[2]
	case "r4i1":
		return &q.Results[3].I[0]
	case "r4i2":
		return &q.Results[3].I[1]
	case "r4i3":
		return &q.Results[3].I[2]
	case "r5i1":
		return &q.Results[4].I[0]
	case "r5i2":
		return &q.Results[4].I[1]
	case "r5i3":
		return &q.Results[4].I[2]
	case "r6i1":
		return &q.Results[5].I[0]
	case "r6i2":
		return &q.Results[5].I[1]
	case "r6i3":
		return &q.Results[5].I[2]
	case "r1c1":
		return &q.Results[0].C[0]
	case "r1c2":
		return &q.Results[0].C[1]
	case "r1c3":
		return &q.Results[0].C[2]
	case "r2c1":
		return &q.Results[1].C[0]
	case "r2c2":
		return &q.Results[1].C[1]
	case "r2c3":
		return &q.Results[1].C[2]
	case "r3c1":
		return &q.Results[2].C[0]
	case "r3c2":
		return &q.Results[2].C[1]
	case "r3c3":
		return &q.Results[2].C[2]
	case "r4c1":
		return &q.Results[3].C[0]
	case "r4c2":
		return &q.Results[3].C[1]
	case "r4c3":
		return &q.Results[3].C[2]
	case "r5c1":
		return &q.Results[4].C[0]
	case "r5c2":
		return &q.Results[4].C[1]
	case "r5c3":
		return &q.Results[4].C[2]
	case "r6c1":
		return &q.Results[5].C[0]
	case "r6c2":
		return &q.Results[5].C[1]
	case "r6c3":
		return &q.Results[5].C[2]
	case "unl1":
		return &q.Results[0].Unl
	case "unl2":
		return &q.Results[1].Unl
	case "unl3":
		return &q.Results[2].Unl
	case "unl4":
		return &q.Results[3].Unl
	case "unl5":
		return &q.Results[4].Unl
	case "unl6":
		return &q.Results[5].Unl
	}
	return nil
}

// Get returns a numeric column.
func (q *QuestDetails) Get(tag Tag) int { return getInt(q.ptr(tag)) }

// Set sets a numeric column (it always reports success, like the original).
func (q *QuestDetails) Set(tag Tag, v int) bool {
	setInt(q.ptr(tag), v)
	return true
}

// Compare compares a numeric column.
func (q *QuestDetails) Compare(o *QuestDetails, tag Tag) int {
	a := q.ptr(tag)
	if a == nil {
		return 0
	}
	return cmp(getInt(a), getInt(o.ptr(tag)))
}

var questTextTags = map[Tag]int{"name": 0, "summ": 1, "desc": 2, "welc": 3, "rmdr": 4, "altm": 5,
	"rep1": 6, "rep2": 7, "rep3": 8, "rep4": 9, "rep5": 10, "rep6": 11}

// Text returns a text column, or "qtyp" as a name.
func (q *QuestDetails) Text(tag Tag) string {
	if i, ok := questTextTags[tag]; ok {
		return *q.strings()[i]
	}
	if tag == "qtyp" {
		switch q.Req.Qtyp {
		case 0:
			return "Final"
		case 1:
			return "Farm"
		case 2:
			return "PvE"
		case 3:
			return "PvP"
		}
		return fmt.Sprintf("unk %d", q.Req.Qtyp)
	}
	return ""
}

// SetText sets a text column.
func (q *QuestDetails) SetText(tag Tag, v string) bool {
	if i, ok := questTextTags[tag]; ok {
		*q.strings()[i] = v
		return true
	}
	return false
}

// stripCrLf removes trailing CR LF pairs (String::stripCrLf).
func stripCrLf(s string) string {
	for strings.HasSuffix(s, "\r\n") {
		s = s[:len(s)-2]
	}
	return s
}

// cleanText trims trailing CR LF and spaces (stripCrLf, stripRight, stripCrLf).
func cleanText(s string) string { return stripCrLf(strings.TrimRight(stripCrLf(s), " ")) }

func (q *QuestDetails) cleanStrings() {
	for _, p := range q.strings() {
		*p = strings.ReplaceAll(cleanText(*p), "\x19", "'")
	}
}

// Read reads a quest (QuestDetails::read).
func (q *QuestDetails) Read(r *binstream.Reader, f Format, opt NpcOptions) {
	switch f {
	case File:
		q.QuestID = r.Int16()
		q.Name = r.String(4, r.CodePage)
		q.Summary = r.String(4, r.CodePage)
		q.Req.read(r)
		if q.Version&0xFFF < 0x600 {
			for k := 0; k < 3; k++ {
				q.Results[k].read(r, q.Version)
			}
			for k := 0; k < 3; k++ {
				q.Replies[k] = r.String(4, r.CodePage)
			}
		} else {
			for k := 0; k < 6; k++ {
				q.Results[k].read(r, q.Version)
				q.Replies[k] = r.String(4, r.CodePage)
			}
		}
		q.Desc = r.String(4, r.CodePage)
		q.Welcome = r.String(4, r.CodePage)
		q.Reminder = r.String(4, r.CodePage)
		q.AltMessage = r.String(4, r.CodePage)
	case Flnt:
		q.QuestID = r.Int16()
		q.Req.read(r)
		for k := range q.Results {
			q.Results[k].read(r, q.Version)
		}
		if opt.CleanCR {
			q.cleanStrings()
		}
		return
	case Ltxt:
		for _, p := range q.ltxtStrings() {
			*p = r.String(4, r.CodePage)
		}
	}
	for _, p := range q.strings() {
		*p = unescape(*p)
	}
	if opt.CleanCR {
		q.cleanStrings()
	}
}

// Write writes a quest (QuestDetails::write).
func (q *QuestDetails) Write(w *binstream.Writer, f Format) {
	cp := w.CodePage
	switch f {
	case File:
		w.Int16(q.QuestID)
		w.String(escape(q.Name), 4, cp)
		w.String(escape(q.Summary), 4, cp)
		q.Req.write(w)
		if q.Version&0xFFF < 0x600 {
			for k := 0; k < 3; k++ {
				q.Results[k].write(w, q.Version)
			}
			for k := 0; k < 3; k++ {
				w.String(escape(q.Replies[k]), 4, cp)
			}
		} else {
			for k := 0; k < 6; k++ {
				q.Results[k].write(w, q.Version)
				w.String(escape(q.Replies[k]), 4, cp)
			}
		}
		for _, s := range []string{q.Desc, q.Welcome, q.Reminder, q.AltMessage} {
			w.String(escape(s), 4, cp)
		}
	case Ltxt:
		for _, p := range q.ltxtStrings() {
			w.String(escape(*p), 4, cp)
		}
	case Flnt:
		w.Int16(q.QuestID)
		q.Req.write(w)
		for k := range q.Results {
			q.Results[k].write(w, q.Version)
		}
	}
}

// Line exports the quest ("----" with ';' or "tabs") (toText).
func (q *QuestDetails) Line(format string) (string, bool) {
	sep := "\t"
	switch format {
	case "----":
		sep = ";"
	case "tabs":
	default:
		return "", false
	}
	return fmt.Sprintf("%d%s", q.QuestID, sep) + q.Req.text(sep), true
}

// Quests is the list of quests of NpcQuest.SData (QuestsDetails).
type Quests struct {
	Quests  []*QuestDetails
	Version int
}

// read reads the quests, trying the EP5 then the EP6 layout for .SData
// files (QuestsDetails::read).
func (s *Quests) read(r *binstream.Reader, f Format, opt NpcOptions) {
	if f == Ltxt {
		n := r.Int()
		for k := 0; k < n; k++ {
			if k < len(s.Quests) {
				s.Quests[k].Read(r, Ltxt, opt)
			}
		}
		return
	}
	saved := r.Pos
	versions := []int{0x500, 0x600}
	if f != File {
		versions = versions[1:]
	}
	var last error
	for _, v := range versions {
		s.Quests = nil
		s.Version = v
		if f == Flnt {
			s.Version |= 0x2000
		}
		r.Pos = saved
		last = binstream.Try(func() {
			n := r.Int()
			for i := 0; i < n; i++ {
				q := NewQuest(s.Version)
				q.Read(r, f, opt)
				s.Quests = append(s.Quests, q)
			}
		})
		if last == nil {
			return
		}
	}
	s.Version = 0
	binstream.Fail("quests: %v", last)
}

func (s *Quests) write(w *binstream.Writer, f Format) {
	w.Uint32(uint32(len(s.Quests)))
	for _, q := range s.Quests {
		q.Write(w, f)
	}
}
