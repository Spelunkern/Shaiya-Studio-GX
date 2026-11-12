package sdata

// getInt reads a numeric field through a pointer returned by a ptr(tag)
// method, with the sign of its C++ type.
func getInt(p any) int {
	switch v := p.(type) {
	case *uint8:
		return int(*v)
	case *int8:
		return int(*v)
	case *uint16:
		return int(*v)
	case *int16:
		return int(*v)
	case *uint32:
		return int(*v)
	case *int32:
		return int(*v)
	case *int:
		return *v
	case *bool:
		if *v {
			return 1
		}
	}
	return 0
}

// setInt stores v truncated to the field size (the C++ casts).
func setInt(p any, v int) bool {
	switch f := p.(type) {
	case *uint8:
		*f = uint8(v)
	case *int8:
		*f = int8(v)
	case *uint16:
		*f = uint16(v)
	case *int16:
		*f = int16(v)
	case *uint32:
		*f = uint32(v)
	case *int32:
		*f = int32(v)
	case *int:
		*f = v
	case *bool:
		*f = v != 0
	default:
		return false
	}
	return true
}

// ItemIconSheet returns the icon sheet of an item type (itemIconSheet).
func ItemIconSheet(t uint8) uint8 {
	switch t {
	case 1, 45:
		return 1
	case 2, 46:
		return 2
	case 3, 47:
		return 3
	case 4, 48:
		return 4
	case 5, 49, 50:
		return 5
	case 6, 51, 52:
		return 6
	case 7, 53, 54:
		return 7
	case 8, 55, 56:
		return 8
	case 9, 57:
		return 9
	case 10, 58:
		return 10
	case 11, 59:
		return 11
	case 12, 60, 61:
		return 12
	case 13, 62, 63:
		return 13
	case 14, 64:
		return 14
	case 15, 65:
		return 15
	case 16, 66, 72:
		return 16
	case 17, 67, 73:
		return 17
	case 18, 68, 74:
		return 18
	case 19, 69:
		return 19
	case 20, 70, 76:
		return 20
	case 21, 71, 77:
		return 21
	case 22, 37:
		return 22
	case 23, 96:
		return 23
	case 24:
		return 24
	case 25:
		return 25
	case 30, 95, 98:
		return 30
	case 31, 81, 87:
		return 31
	case 32, 82, 88:
		return 32
	case 33, 83, 89:
		return 33
	case 34, 84:
		return 34
	case 35, 85, 91:
		return 35
	case 36, 86, 92:
		return 36
	case 39:
		return 39
	case 40, 97:
		return 40
	}
	return t
}

// ItemIconBase returns the base icon id of an item type (itemIconBase).
func ItemIconBase(t uint8) int {
	switch {
	case t == 25 || t == 38 || t == 41 || t == 42 || t == 43 || t == 44 || t == 94:
		return 25000
	case t-27 <= 2:
		return 27000
	case t-78 <= 2:
		return 78000
	case t == 100 || t == 101:
		return 100000
	case t == 120 || t == 150:
		return 1000 * int(t)
	}
	return 1000 * int(ItemIconSheet(t))
}
