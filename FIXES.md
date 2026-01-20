# UX Fixes Applied - 2026-01-20

## Issues Fixed

### ✅ 1. Keybindings Not Shown in Help

**Problem:** The 'J' and 'y' keybindings worked but weren't displayed in the help panel.

**Solution:**
- Modified `ShortHelp()` and `FullHelp()` methods in `containers.go`
- When detail pane is focused, display detail-specific keybindings:
  - `↑/k` - Scroll up
  - `↓/j` - Scroll down
  - `J` - Toggle JSON/YAML
  - `y` - Copy to clipboard
  - `tab` - Switch focus
- When list is focused, show standard list keybindings

**Files Changed:**
- `internal/ui/containers/containers.go` (lines 421-448, 557-575)

---

### ✅ 2. Scroll Position Not Saved Per Container

**Problem:** Scroll position was shared across all containers instead of being tracked individually.

**Root Cause:** The `ViewportPane.SetSize()` method in `splitview.go` was creating a NEW viewport instance every time it was called, destroying all content and scroll position state.

**Solution:**

1. **Fixed Critical Bug in ViewportPane.SetSize():**
   ```go
   // BEFORE (BUG):
   func (v *ViewportPane) SetSize(w, h int) {
       v.Viewport = viewport.New(viewport.WithWidth(w), viewport.WithHeight(h)) // Creates NEW viewport!
   }
   
   // AFTER (FIXED):
   func (v *ViewportPane) SetSize(w, h int) {
       v.Viewport.SetWidth(w)
       v.Viewport.SetHeight(h) // Updates existing viewport
   }
   ```

2. **Added Deferred Scroll Restoration:**
   - Created `MsgRestoreScroll` message type
   - After setting content, send `MsgRestoreScroll` on next tick
   - This allows viewport to process content before restoring scroll position

3. **Removed Premature Save in refreshInspectionContent():**
   - Was saving scroll position when refreshing (format toggle)
   - This was overwriting the saved position incorrectly
   - Now only saves when switching containers, restores when loading content

**Files Changed:**
- `internal/ui/components/splitview.go` (lines 56-58)
- `internal/ui/containers/containers.go` (lines 31-33, 388-394, 473-498)

**How It Works Now:**
1. User scrolls in container A to position 100
2. User switches to container B
   - `saveScrollPosition()` saves A's position (100) in map
   - New content loads for B
   - `MsgRestoreScroll` restores B's saved position (or 0 if first visit)
3. User switches back to container A
   - B's position saved
   - A's content loads
   - A's position (100) restored from map

---

### ✅ 3. Viewport Going Off Screen (No MaxHeight)

**Problem:** The informational panel content extended beyond the screen, making the bottom border invisible.

**Root Cause:** Same as issue #2 - the viewport was being recreated on every resize, which prevented proper height constraints.

**Solution:**
- Fixed by correcting `ViewportPane.SetSize()` to use `SetWidth()` and `SetHeight()`
- The viewport's Height property now properly constrains the content
- Viewport automatically handles scrolling when content exceeds height
- Border remains visible at all times

**Files Changed:**
- `internal/ui/components/splitview.go` (lines 56-58)

**How It Works:**
- SplitView calculates available height: `detailLayout.Height - 2` (for border)
- Calls `SetSize(w, h)` on viewport with calculated height
- Viewport now properly respects height constraint
- Content scrolls within the viewport
- Border always visible at calculated height

---

## Testing Recommendations

### Test Case 1: Keybindings Visibility
1. Start application
2. Press Tab to focus detail pane
3. Press `?` to show help
4. **Expected:** Help shows `J` and `y` keybindings with descriptions

### Test Case 2: Scroll Position Per Container
1. Select container A
2. Scroll down 10 lines (press `j` 10 times)
3. Press Tab, then arrow key to select container B
4. **Expected:** Container B starts at top (scroll position = 0)
5. Scroll down 5 lines in container B
6. Switch back to container A
7. **Expected:** Container A is still scrolled to line 10

### Test Case 3: Format Toggle Preserves Scroll
1. Select any container
2. Scroll down 20 lines
3. Press Tab to focus detail, then press `J` to toggle format
4. **Expected:** Scroll position remains at line 20

### Test Case 4: Viewport Height Constraint
1. Select a container with very long configuration (1000+ lines)
2. **Expected:** 
   - Bottom border of detail pane is visible
   - Content scrolls within the pane
   - Border doesn't move off screen

### Test Case 5: Window Resize
1. Select a container and scroll to middle
2. Resize terminal window (smaller then larger)
3. **Expected:**
   - Scroll position is maintained
   - Content reflows to new width
   - No crashes or content loss

---

## Additional Notes

### The ViewportPane Bug Impact

The bug in `ViewportPane.SetSize()` was causing multiple issues:
- Every window resize created a new viewport → lost all scroll state
- Every resize lost content → had to reload
- Scroll position tracking was fighting against viewport recreation
- Height constraints weren't properly applied

By fixing this single bug, we resolved issues #2 and #3 simultaneously.

### Scroll Position Timing

The use of `MsgRestoreScroll` on the next tick is necessary because:
1. `SetContent()` internally resets viewport's YOffset to 0
2. We need to restore scroll AFTER the viewport processes the new content
3. Doing it on the same tick doesn't work reliably
4. Using a deferred message ensures proper ordering

### Code Quality Improvements

This fix also improves:
- **Performance:** No longer recreating viewport on every resize
- **Reliability:** Viewport state is preserved across window operations
- **Maintainability:** Clearer separation between saving and restoring scroll

---

## Files Modified Summary

1. **`internal/ui/components/splitview.go`**
   - Fixed ViewportPane.SetSize() to update existing viewport instead of creating new one

2. **`internal/ui/containers/containers.go`**
   - Added MsgRestoreScroll message type
   - Updated ShortHelp() to show detail keybindings when detail is focused
   - Updated FullHelp() to show detail keybindings when detail is focused
   - Modified MsgContainerInspection handler to send MsgRestoreScroll
   - Added MsgRestoreScroll handler to restore scroll position
   - Fixed refreshInspectionContent() to not save scroll prematurely

**Total Lines Changed:** ~50 lines across 2 files
**Build Status:** ✅ Success
**Breaking Changes:** None (bug fixes only)

---

*Applied: 2026-01-20*
*Status: Complete and Ready for Testing*
