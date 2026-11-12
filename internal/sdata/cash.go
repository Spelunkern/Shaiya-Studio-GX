package sdata

import (
	"fmt"
	"strings"

	"github.com/Spelunkern/Shaiya-Studio-GX/internal/binstream"
)

// BagNames are the cash shop tab names (g_bagNames).
var BagNames = []string{"EXP & More", "Buffs", "Character & Fun", "Gear & Upgrades", "Costumes", "Pets"}

// BagName returns the name of bag n (1-based), "Unknown n" past the known
// ones, or "" for n < 1 (g3_bagName).
func BagName(n int) string {
	if n < 1 {
		return ""
	}
	if n > len(BagNames) {
		return fmt.Sprintf("Unknown %d", n)
	}
	return BagNames[n-1]
}

// CashProduct is a row of Cash.SData / DBItemSellData (CashProduct).
type CashProduct struct {
	Version   int
	Row       int32
	Bag       int32
	Icon      int32
	Cost      int32
	Cost2     int32
	ItemID    [24]int32
	ItemCount [24]uint8
	Name      string
	Code      string
	Desc      string
	Title     string // "<name> - <cost> Pts"
	ItemTotal int    // number of non-zero item ids
}

// NewCashProduct returns a reset product in bag 1.
func NewCashProduct(version int, row int32) *CashProduct {
	return &CashProduct{Version: version, Row: row, Bag: 1}
}

func (p *CashProduct) updateTitle() { p.Title = fmt.Sprintf("%s - %d Pts", p.Name, p.Cost) }

// Reset clears the product (bag 1).
func (p *CashProduct) Reset() {
	p.Bag, p.Icon, p.Cost, p.Cost2 = 1, 0, 0, 0
	p.ItemID = [24]int32{}
	p.ItemCount = [24]uint8{}
	p.ItemTotal = 0
}

// Assign copies src (operator=), keeping the version.
func (p *CashProduct) Assign(src *CashProduct) bool {
	v := p.Version
	*p = *src
	p.Version = v
	return true
}

// IconID returns the icon (getId).
func (p *CashProduct) IconID() int { return int(p.Icon) }

// Text returns "name", "code", "desc" or "ttle".
func (p *CashProduct) Text(tag Tag) string {
	switch tag {
	case "name":
		return p.Name
	case "code":
		return p.Code
	case "desc":
		return p.Desc
	case "ttle":
		return p.Title
	}
	return ""
}

// SetText sets the name (updating the title), code or description.
func (p *CashProduct) SetText(tag Tag, v string) bool {
	switch tag {
	case "name":
		p.Name = v
		p.updateTitle()
	case "code":
		p.Code = v
	case "desc":
		p.Desc = v
	default:
		return false
	}
	return true
}

func slot(tag Tag, prefix string) int {
	if len(tag) == 4 && string(tag[:2]) == prefix {
		var n int
		if _, err := fmt.Sscanf(string(tag[2:]), "%02d", &n); err == nil && n >= 1 && n <= 24 {
			return n - 1
		}
	}
	return -1
}

// Get returns a numeric column.
func (p *CashProduct) Get(tag Tag) int {
	switch tag {
	case "row ":
		return int(p.Row)
	case "type":
		return int(p.Bag)
	case "icon":
		return int(p.Icon)
	case "cost":
		return int(p.Cost)
	case "cst2":
		return int(p.Cost2)
	case "#cnt":
		return p.ItemTotal
	}
	if i := slot(tag, "it"); i >= 0 {
		return int(p.ItemID[i])
	}
	if i := slot(tag, "ic"); i >= 0 {
		return int(p.ItemCount[i])
	}
	return 0
}

// Set sets the cost (the only editable numeric column; it updates the title).
func (p *CashProduct) Set(tag Tag, v int) bool {
	if tag == "cost" {
		p.Cost = int32(v)
		p.updateTitle()
		return true
	}
	return false
}

// Compare compares two products on tag.
func (p *CashProduct) Compare(o *CashProduct, tag Tag) int {
	switch tag {
	case "row ":
		return cmp(p.Row, o.Row)
	case "type":
		return cmp(p.Bag, o.Bag)
	case "icon":
		return cmp(p.Icon, o.Icon)
	case "cost":
		return cmp(p.Cost, o.Cost)
	case "name":
		return compareLogical(p.Name, o.Name)
	case "code":
		return compareLogical(p.Code, o.Code)
	case "desc":
		return compareLogical(p.Desc, o.Desc)
	case "#cnt":
		return cmp(p.ItemTotal, o.ItemTotal)
	}
	return 0
}

func (p *CashProduct) countItems() {
	p.ItemTotal = 0
	for _, id := range p.ItemID {
		if id != 0 {
			p.ItemTotal++
		}
	}
}

// Read reads a row (CashProduct::read).
func (p *CashProduct) Read(r *binstream.Reader, f Format) {
	switch f {
	case File:
		p.Row = r.Int32()
		p.Bag = r.Int32()
		p.Icon = r.Int32()
		p.Cost = r.Int32()
		for i := range p.ItemID {
			p.ItemID[i] = r.Int32()
			p.ItemCount[i] = r.Byte()
		}
		p.countItems()
		p.Name = r.String(4, r.CodePage)
		p.Code = r.String(4, r.CodePage)
		p.Desc = r.String(4, r.CodePage)
		p.updateTitle()
		p.Desc = unescape(p.Desc)
	case Flnt:
		p.Code = r.String(1, 0)
		p.Row = int32(r.Int64())
		p.Cost = int32(r.Int64())
		if p.Version&0xFFF >= 0x700 {
			p.Cost2 = int32(r.Int64())
		}
		for i := range p.ItemID {
			p.ItemID[i] = int32(r.Int64())
			p.ItemCount[i] = uint8(r.Int64())
		}
		p.countItems()
		p.Bag = int32(r.Int64())
		p.Icon = int32(r.Int64())
	case Ltxt:
		p.Name = r.String(4, -1)
		p.updateTitle()
		p.Desc = unescape(r.String(4, -1))
	}
}

// Write writes a row (CashProduct::write). In the .SData format the
// strings carry an extra NUL, like the original.
func (p *CashProduct) Write(w *binstream.Writer, f Format) {
	switch f {
	case File:
		w.Int32(p.Row)
		w.Int32(p.Bag)
		w.Int32(p.Icon)
		w.Int32(p.Cost)
		for i := range p.ItemID {
			w.Int32(p.ItemID[i])
			w.Byte(p.ItemCount[i])
		}
		w.String(p.Name+"\x00", 4, w.CodePage)
		w.String(p.Code+"\x00", 4, w.CodePage)
		w.String(escape(p.Desc)+"\x00", 4, w.CodePage)
	case Flnt:
		w.String(p.Code, 1, 0)
		w.Int64(int64(p.Row))
		w.Int64(int64(p.Cost))
		if p.Version&0xFFF >= 0x700 {
			w.Int64(int64(p.Cost2))
		}
		for i := range p.ItemID {
			w.Int64(int64(p.ItemID[i]))
			w.Int64(int64(p.ItemCount[i]))
		}
		w.Int64(int64(p.Bag))
		w.Int64(int64(p.Icon))
	case Ltxt:
		w.Int64(int64(p.Row))
		w.String(p.Name, 4, -1)
		w.String(escape(p.Desc), 4, -1)
	}
}

// Line exports a row as CSV or SQL (toText).
func (p *CashProduct) Line(format string) (string, bool) {
	switch format {
	case TextSQL:
		cols := "[ProductName],[ProductCode],[BuyCost]"
		vals := sqlQuote(p.Name) + "," + sqlQuote(p.Code) + "," + i32s(p.Cost)
		for i := 1; i <= 24; i++ {
			cols += fmt.Sprintf(",[ItemID%d],[ItemCount%d]", i, i)
			vals += fmt.Sprintf(",%d,%d", p.ItemID[i-1], p.ItemCount[i-1])
		}
		return "INSERT INTO [PS_GameDefs].[dbo].[ProductList] (" + cols + ")) values (" + vals + ")\r\n", true
	case TextCSV:
		var b strings.Builder
		fmt.Fprintf(&b, "%d;%d;%d;%d;%s;%s;%s", p.Row, p.Bag, p.Icon, p.Cost, p.Code, p.Name, escape(p.Desc))
		for i := range p.ItemID {
			fmt.Fprintf(&b, ";%d;%d", p.ItemID[i], p.ItemCount[i])
		}
		b.WriteString("\r\n")
		return b.String(), true
	}
	return "", false
}

// CashProducts is Cash.SData / DBItemSellData / DBItemSellText
// (CashProducts).
type CashProducts struct {
	Products []*CashProduct
	Version  int
	DataHead []byte
	TextHead []byte
}

// NewCashProducts returns an empty table.
func NewCashProducts(version int) *CashProducts { return &CashProducts{Version: version} }

// FilePrefix is the prefix of the DB text files.
func (s *CashProducts) FilePrefix() string { return `BinarySData\DBItemSellText_` }

// Bags lists the distinct bags in row order (sub_420770).
func (s *CashProducts) Bags() []int32 {
	var bags []int32
	for _, p := range s.Products {
		found := false
		for _, b := range bags {
			if b == p.Bag {
				found = true
				break
			}
		}
		if !found {
			bags = append(bags, p.Bag)
		}
	}
	return bags
}

// FindCode returns the product with a code (sub_420800).
func (s *CashProducts) FindCode(code string) *CashProduct {
	for _, p := range s.Products {
		if strings.EqualFold(p.Code, code) {
			return p
		}
	}
	return nil
}

// SetVersion sets the version of the table and its rows.
func (s *CashProducts) SetVersion(v int) {
	s.Version = v
	for _, p := range s.Products {
		p.Version = v
	}
}

// Read loads the table (CashProducts::read).
func (s *CashProducts) Read(data []byte, cp int, f Format) error {
	return binstream.Try(func() {
		r := binstream.NewReader(data, cp)
		if f == Ltxt {
			s.TextHead = r.Copy(128)
			if r.Uint32() != 2 {
				failDB()
			}
			checkSignature(r, "ProductName")
			checkSignature(r, "text")
			n := int(r.Uint32())
			for ; n > 0; n-- {
				row := int32(r.Int64())
				var found *CashProduct
				for _, p := range s.Products {
					if p.Row == row {
						found = p
						break
					}
				}
				if found != nil {
					found.Read(r, Ltxt)
				} else {
					r.String(4, -1)
					r.String(4, -1)
				}
			}
			return
		}
		s.Products = nil
		if f == Flnt {
			s.DataHead = r.Copy(128)
			v := r.Uint32()
			switch v {
			case 54:
				s.Version = 0x1700
			case 53:
				s.Version = 0x1600
			default:
				failDB()
			}
			checkSignature(r, "ProductCode")
			checkSignature(r, "goods_id")
			if v == 53 {
				checkSignature(r, "BuyCost")
			} else {
				checkSignature(r, "Multi_BuyCost1")
				checkSignature(r, "Multi_BuyCost2")
			}
			for i := 1; i <= 24; i++ {
				checkSignature(r, fmt.Sprintf("ItemID%d", i))
				checkSignature(r, fmt.Sprintf("ItemCount%d", i))
			}
			checkSignature(r, "type")
			checkSignature(r, "icon")
		}
		n := int(r.Uint32())
		for i := 0; i < n; i++ {
			p := NewCashProduct(s.Version, 0)
			p.Read(r, f)
			s.Products = append(s.Products, p)
		}
	})
}

// Write serializes the table (CashProducts::write).
func (s *CashProducts) Write(cp int, f Format) ([]byte, error) {
	w := binstream.NewWriter(cp)
	switch f {
	case Flnt:
		w.Put(head(s.DataHead))
		v7 := s.Version&0xFFF >= 0x700
		if v7 {
			w.Uint32(54)
		} else {
			w.Uint32(53)
		}
		writeColumn(w, "ProductCode")
		writeColumn(w, "goods_id")
		if v7 {
			writeColumn(w, "Multi_BuyCost1")
			writeColumn(w, "Multi_BuyCost2")
		} else {
			writeColumn(w, "BuyCost")
		}
		for i := 1; i <= 24; i++ {
			writeColumn(w, fmt.Sprintf("ItemID%d", i))
			writeColumn(w, fmt.Sprintf("ItemCount%d", i))
		}
		writeColumn(w, "type")
		writeColumn(w, "icon")
	case Ltxt:
		w.Put(head(s.TextHead))
		w.Uint32(2)
		writeColumn(w, "ProductName")
		writeColumn(w, "text")
	}
	w.Uint32(uint32(len(s.Products)))
	for _, p := range s.Products {
		p.Write(w, f)
	}
	return w.Bytes(), nil
}

// Text exports the table as CSV or SQL (toText / toSql).
func (s *CashProducts) Text(format string) string {
	var b strings.Builder
	switch format {
	case TextCSV:
		b.WriteString("RowID;Bag;Icon;BuyCost;ProductCode;ProductName;ProductDesc")
		for i := 1; i <= 24; i++ {
			fmt.Fprintf(&b, ";ItemID%d;ItemCount%d", i, i)
		}
		b.WriteString("\r\n")
	case TextSQL:
		b.WriteString("TRUNCATE TABLE [PS_GameDefs].[dbo].[ProductList]\r\n")
	default:
		return ""
	}
	for _, p := range s.Products {
		line, _ := p.Line(format)
		b.WriteString(line)
	}
	return b.String()
}
