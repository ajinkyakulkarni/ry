package main

import (
	"regexp"
	"strings"
)

var (
	last_search_buffer    *buffer = nil
	last_search                   = ""
	last_search_index             = 0
	last_search_highlight         = false
	last_search_results           = []*location{}
)

func init_search() {
	bind("normal", k("/"), handle_search_start)
	bind("normal", k("N"), handle_search_prev)
	bind("normal", k("n"), handle_search_next)
	bind("normal", k("*"), handle_search_search_work_under_cursor)
	bind("normal", k("SPC n"), func(vt *view_tree, b *buffer, kl *key_list) {
		search_clear()
	})

	add_command("clearsearch", func(args []string) {
		search_clear()
	})
	add_alias("cs", "clearsearch")

	hook_buffer("modified", func(b *buffer) {
		// Update search result indexes on buffer changes
		if last_search != "" && b == last_search_buffer {
			search_find_matches(b, last_search)
			last_search_index = len(last_search_results) - 1
		}
	})
}

func search_clear() {
	last_search_highlight = false
	highlight_buffer(current_view_tree.leaf.buf)
}

func search_find_matches(b *buffer, search string) {
	last_search = search
	re := regexp.MustCompile(regexp.QuoteMeta(search))

	last_search_buffer = b
	last_search_results = []*location{}
	for i, line := range b.data {
		idxs := re.FindAllStringIndex(string(line), -1)
		for _, idx := range idxs {
			last_search_results = append(last_search_results, new_location(i, idx[0]))
		}
	}
}

func search_start(b *buffer, search string) {
	if len(search) > 0 {
		search_find_matches(b, search)

		// TODO start index at first match after cursor
		last_search_index = len(last_search_results) - 1
		search_next(b)
	}
}

func handle_search_start(vt *view_tree, b *buffer, kl *key_list) {
	prompt("/", noop_complete, func(args []string) {
		search_start(b, strings.Join(args, " "))
	})
}

func search_prev(b *buffer) {
	if len(last_search_results) == 0 {
		message("No search result.")
		return
	}
	if last_search_index == 0 {
		last_search_index = len(last_search_results) - 1
	} else {
		last_search_index--
	}
	last_search_highlight = true
	highlight_buffer(current_view_tree.leaf.buf)
	loc := last_search_results[last_search_index]
	b.move_to(loc.char, loc.line)

}

func handle_search_prev(vt *view_tree, b *buffer, kl *key_list) {
	search_prev(b)
}

func search_next(b *buffer) {
	if len(last_search_results) == 0 {
		message("No search result.")
		return
	}
	if last_search_index+1 == len(last_search_results) {
		last_search_index = 0
	} else {
		last_search_index++
	}
	last_search_highlight = true
	highlight_buffer(current_view_tree.leaf.buf)
	loc := last_search_results[last_search_index]
	b.move_to(loc.char, loc.line)

}

func handle_search_next(vt *view_tree, b *buffer, kl *key_list) {
	search_next(b)
}

func search_highlight(b *buffer, l, c int) int {
	if !last_search_highlight {
		return 0
	}
	if last_search_buffer != b {
		return 0
	}
	for _, loc := range last_search_results {
		if l == loc.line && c == loc.char {
			return len(last_search)
		}
	}
	return 0
}

func handle_search_search_work_under_cursor(vt *view_tree, b *buffer, kl *key_list) {
	search_start(b, string(b.word_under_cursor()))
}
