# Footguns and issues to fix

- stuiview.Selection is a map of booleans (row identifiers to true/false), but acually only the *presence* of a key in this map is used for selection, highlight and command logic. Should be fixed, probably using some kind of custom Selection struct that has simpler access functions and keeps its internal state cleanly.
- Notifications and live updates across in many places (e.g. selection counter, object count counter in a table) are implemented in various ways like callbacks, direct access, `tview` functions. Should be centralized to a channel-based notification/message system so individual text data can be updated with less hacks.

