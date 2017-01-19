package db

type varlist struct {
	list    []*varentry
	indexes map[string]int
}

func newvarlist() *varlist {
	return &varlist{
		indexes: make(map[string]int),
	}
}

type varentry struct {
	value string
	next  *varentry
	prev  *varentry
}

func (vl *varlist) insertAfter(value, mark string) {
	newentry := &varentry{
		value: value,
	}
	if idx, found := vl.indexes[mark]; found {
		mark := vl.list[idx]
		// if mark's next is nil, we just add
		if mark.next == nil {
			mark.next = newentry
			newentry.prev = mark
			vl.list = append(vl.list, newentry)
			vl.indexes[value] = len(vl.list) - 1
			return
		}
		// else, save the tmp
		oldnext := mark.next
		// newentry is here
		mark.next = newentry
		oldnext.prev = newentry
		newentry.next = oldnext
		newentry.prev = mark
		vl.list = append(vl.list[:idx+1], append([]*varentry{newentry}, vl.list[idx+1:]...)...)
		vl.rebuildindex()
	}
}

func (vl *varlist) dumplist() []string {
	var ret = make([]string, len(vl.list))
	for idx, val := range vl.list {
		ret[idx] = val.value
	}
	return ret
}

func (vl *varlist) rebuildindex() {
	for idx, val := range vl.list {
		vl.indexes[val.value] = idx
	}
}

func (vl *varlist) has(val string) bool {
	_, found := vl.indexes[val]
	return found
}

func (vl *varlist) remove(value string) {
	if idx, found := vl.indexes[value]; found {
		entry := vl.list[idx]
		if idx == 0 {
			vl.list = vl.list[1:]
			if entry.next != nil {
				entry.next.prev = nil
			}
			return
		}
		entry.prev.next = entry.next
		entry.next.prev = entry.prev
		vl.list = append(vl.list[:idx], vl.list[idx+1:]...)
	}
}

// take [value] and all of its subsequent children up until the mark and append it after mark
func (vl *varlist) moveAfter(value, mark string) {
	var last_link *varentry

	value_idx := vl.indexes[value]
	mark_idx := vl.indexes[mark]
	value_entry := vl.list[value_idx]
	mark_entry := vl.list[mark_idx]

	if mark_entry.prev == nil { // first entry
		last_link = vl.list[len(vl.list)-1]
		last_link.next = mark_entry.next
		if mark_entry.next != nil {
			mark_entry.next.prev = last_link
		}
		mark_entry.next = value_entry
	} else {
		// this is the 'last' element to move connected to the value list
		last_link = mark_entry.prev
		last_link.next = mark_entry.next
		// connect the mark to what came before value
		mark_entry.prev = value_entry.prev
		if mark_entry.next != nil {
			mark_entry.next.prev = last_link
		}
	}
	mark_entry.next = value_entry
	value_entry.prev = mark_entry

	if value_idx < mark_idx {
		vl.list = append(vl.list[:value_idx], append(vl.list[mark_idx:], vl.list[value_idx:mark_idx]...)...)
	} else if value_idx > mark_idx {
		vl.list = append(vl.list[:mark_idx+1], append(vl.list[value_idx:], vl.list[mark_idx+1:value_idx]...)...)
	}
	vl.rebuildindex()
}

func (vl *varlist) pushBack(value string) {
	newentry := &varentry{
		value: value,
	}
	if len(vl.list) == 0 {
		vl.list = append(vl.list, newentry)
		vl.indexes[value] = 0
		return
	}
	last_entry := vl.list[len(vl.list)-1]
	vl.list = append(vl.list, newentry)
	vl.indexes[value] = len(vl.list) - 1
	last_entry.next = newentry
	newentry.prev = last_entry
}
