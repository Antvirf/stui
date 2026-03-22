package view

import (
	"fmt"
	"sync"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/logger"
	"github.com/antvirf/stui/internal/model"
	"github.com/antvirf/stui/internal/state"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	COMMAND_PAGE = "command_modal"
)

type App struct {
	App                 *tview.Application
	Pages               *tview.Pages
	PagesContainer      *tview.Flex  // Container for pages with border title
	startTime           time.Time    // Start time of the application
	CurrentTableView    *tview.Table // Points to current table view
	FirstRenderComplete bool
	StateStore          *state.StuiState // State of filter(s) for caching

	// Base app components
	HeaderGrid              *tview.Grid
	HeaderGridInnerContents *tview.Grid
	MainFlex                *tview.Flex
	FooterGrid              *tview.Grid

	// Footer
	HeaderLineOne   *tview.TextView
	HeaderLineTwo   *tview.TextView
	HeaderLineThree *tview.TextView

	// Current tab indicators
	TabNodesBox         *tview.TextView
	TabJobsBox          *tview.TextView
	TabSchedulerBox     *tview.TextView
	TabAccountingMgrBox *tview.TextView
	TabAccountingBox    *tview.TextView

	// Dropdown selectors
	PartitionSelector      *tview.DropDown
	SacctMgrEntitySelector *tview.DropDown
	NodeStateSelector      *tview.DropDown
	JobStateSelector       *tview.DropDown
	SortSelector           *tview.DropDown

	// Search state
	SearchBox     *tview.InputField
	SearchActive  bool
	SearchPattern string
	searchTimer   *time.Timer // Timer for debouncing search updates

	// Command modal state
	CommandModalOpen bool

	// Data  and providers
	PartitionsData     *model.TableData
	PartitionsProvider model.DataProvider[*model.TableData]
	NodesProvider      model.DataProvider[*model.TableData]
	JobsProvider       model.DataProvider[*model.TableData]
	SacctMgrProvider   model.DataProvider[*model.TableData]
	SacctProvider      model.DataProvider[*model.TableData]
	SdiagProvider      model.DataProvider[*model.TextData]

	// New style views
	NodesView    *StuiView
	JobsView     *StuiView
	SacctMgrView *StuiView
	SacctView    *StuiView
	SchedView    *tview.TextView // Special case, text only
}

// Initializes a `stui` instance tview Application using the config module
func InitializeApplication(stateStore *state.StuiState) *App {
	application := App{
		startTime:               time.Now(),
		App:                     tview.NewApplication(),
		Pages:                   tview.NewPages(),
		HeaderGridInnerContents: tview.NewGrid(),
		FirstRenderComplete:     false,
		StateStore:              stateStore,
	}

	// Load previous/saved state from state store
	application.LoadStateFromStateStore()

	// Init data providers at start - in parallel, as they all do their first fetch on initialization.
	// Bools set whether to load data at startup. We load data if quickstart is false, OR
	// if the startup pane is that particular pane's provider.

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(6)
	go func() {
		defer wg.Done()
		application.PartitionsProvider = model.NewPartitionsProvider()
	}()
	go func() {
		defer wg.Done()
		application.NodesProvider = model.NewNodesProvider(!config.Quickstart || config.StartPane == config.NODES_PAGE)
	}()
	go func() {
		defer wg.Done()
		application.JobsProvider = model.NewJobsProvider(!config.Quickstart || config.StartPane == config.JOBS_PAGE)
	}()
	go func() {
		defer wg.Done()
		application.SacctProvider = model.NewSacctProvider(!config.Quickstart || config.StartPane == config.SACCT_PAGE)
	}()
	go func() {
		defer wg.Done()
		application.SacctMgrProvider = model.NewSacctMgrProvider(!config.Quickstart || config.StartPane == config.SACCTMGR_PAGE)
	}()
	go func() {
		defer wg.Done()
		application.SdiagProvider = model.NewSdiagProvider(!config.Quickstart || config.StartPane == config.SDIAG_PAGE)
	}()
	wg.Wait()
	logger.Printf("START: Initial data load from scheduler took %d ms", time.Since(start).Milliseconds())
	return &application
}

func (a *App) SetupViews() {
	a.SetupSearchBox()
	a.SetupSortSelector()
	a.SetupPartitionSelector()
	a.SetupNodeStateSelector()
	a.SetupJobStateSelector()
	a.SetupSacctMgrEntitySelector()

	{ // Header lines
		a.HeaderLineOne = tview.NewTextView().
			SetDynamicColors(true).
			SetTextAlign(tview.AlignLeft)

		// Combined status line
		a.HeaderLineTwo = tview.NewTextView().
			SetDynamicColors(true).
			SetTextAlign(tview.AlignLeft).
			SetWrap(true)

		a.HeaderLineThree = tview.NewTextView().
			SetDynamicColors(true).
			SetTextAlign(tview.AlignCenter)

	}

	{ // Current tab boxes
		a.TabNodesBox = tview.NewTextView().
			SetText("(1) Nodes              [scontrol]")
		a.TabJobsBox = tview.NewTextView().
			SetText("(2) Jobs queue         [scontrol]")
		a.TabAccountingBox = tview.NewTextView().
			SetText("(3) Jobs accounting    [sacct]")
		a.TabAccountingMgrBox = tview.NewTextView().
			SetText("(4) Accounting manager [sacctmgr]")
		a.TabSchedulerBox = tview.NewTextView().
			SetText("(5) Scheduler          [sdiag]")

		// If sacct disabled, blank out those rows
		if !config.SacctEnabled {
			a.TabAccountingBox.SetText("")
			a.TabAccountingMgrBox.SetText("")
		}
	}

	// Create a grid for the tabs
	tabGrid := tview.NewGrid().
		AddItem(a.TabNodesBox, FRST_ROW, FRST_COL, 1, 1, 1, 0, false).
		AddItem(a.TabJobsBox, SCND_ROW, FRST_COL, 1, 1, 1, 0, false).
		AddItem(a.TabAccountingBox, THRD_ROW, FRST_COL, 1, 1, 1, 0, false).
		AddItem(a.TabAccountingMgrBox, FRTH_ROW, FRST_COL, 1, 1, 1, 0, false).
		AddItem(a.TabSchedulerBox, FFTH_ROW, FRST_COL, 1, 1, 1, 0, false)

	a.HeaderGrid = tview.NewGrid().
		SetColumns(-1, -2, -1).
		SetBorders(true).
		AddItem(a.HeaderGridInnerContents, FRST_ROW, FRST_COL, 1, 1, 0, 0, false).
		AddItem(
			tview.NewGrid().
				SetRows(-1, -2, -1).
				AddItem(a.HeaderLineOne, FRST_ROW, FRST_COL, 1, 1, 0, 0, false).
				AddItem(a.HeaderLineTwo, SCND_ROW, FRST_COL, 1, 1, 0, 0, false).
				AddItem(a.HeaderLineThree, THRD_ROW, FRST_COL, 1, 1, 0, 0, false),
			FRST_ROW, SCND_COL, 1, 1, 0, 0, false).
		AddItem(tabGrid, FRST_ROW, THRD_COL, 1, 1, 0, 0, false)

	a.PagesContainer = tview.NewFlex().SetDirection(tview.FlexRow)

	a.PagesContainer.AddItem(a.Pages, 0, 30, true).
		SetBorder(true).
		SetBorderStyle(
			tcell.StyleDefault.
				Foreground(pagesBorderColor).
				Background(generalBackgroundColor),
		)

	// Main grid layout, implemented with Flex
	a.MainFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.HeaderGrid, 7, 0, false).
		AddItem(a.PagesContainer, 0, 1, true)

	a.MainFlex.SetBorder(true).
		SetBorderAttributes(tcell.AttrDim)
	a.UpdateAppTitle()
	a.MainFlex.SetTitleAlign(tview.AlignCenter)

	{ // Nodes View
		a.NodesView = NewStuiView(
			"Nodes",
			a.NodesProvider,
			a.PagesContainer.SetTitle,
			a.UpdateHeaderLineTwo,           // errors
			a.UpdateHeaderLineOne,           // data updates notify
			a.copyCellToClipBoard,           // func to run when a data cell is clicked
			a.SortSelector.SetCurrentOption, // func to run when a header row is clicked
			&a.SearchPattern,                // pointer to search string
		)
		a.Pages.AddPage(config.NODES_PAGE, a.NodesView.Grid, true, true)
	}

	{ // Jobs View
		a.JobsView = NewStuiView(
			"Jobs",
			a.JobsProvider,
			a.PagesContainer.SetTitle,
			a.UpdateHeaderLineTwo,           // errors
			a.UpdateHeaderLineOne,           // data updates notify
			a.copyCellToClipBoard,           // func to run when a data cell is clicked
			a.SortSelector.SetCurrentOption, // func to run when a header row is clicked
			&a.SearchPattern,                // pointer to search string
		)
		a.Pages.AddPage(config.JOBS_PAGE, a.JobsView.Grid, true, false)
	}

	{
		// Accounting views - we create these views whether not they will be used.
		// This way we do not need to gate our code everywhere to check for
		// whether it's enabled, just to avoid segfaults.
		a.SacctMgrView = NewStuiView(
			model.SACCTMGR_TABLE_ENTITIES[0], // First type of entity to start with
			a.SacctMgrProvider,
			a.PagesContainer.SetTitle,
			a.UpdateHeaderLineTwo, // errors
			a.UpdateHeaderLineOne, // data updates notify
			//a.copyToClipBoard,               // func to run when a data cell is clicked
			a.ShowModalPopupMinimal,
			a.SortSelector.SetCurrentOption, // func to run when a header row is clicked
			&a.SearchPattern,                // pointer to search string
		)
		a.Pages.AddPage(config.SACCTMGR_PAGE, a.SacctMgrView.Grid, true, false)

		a.SacctView = NewStuiView(
			"Jobs Accounting",
			a.SacctProvider,
			a.PagesContainer.SetTitle,
			a.UpdateHeaderLineTwo,           // errors
			a.UpdateHeaderLineOne,           // data updates notify
			a.copyCellToClipBoard,           // func to run when a data cell is clicked
			a.SortSelector.SetCurrentOption, // func to run when a header row is clicked
			&a.SearchPattern,                // pointer to search string
		)

		a.Pages.AddPage(config.SACCT_PAGE, a.SacctView.Grid, true, false)
	}

	{ // Scheduler View
		a.SchedView = tview.NewTextView()
		a.SchedView.
			SetDynamicColors(true).
			SetScrollable(true).
			SetWrap(false).
			SetTitleAlign(tview.AlignLeft).
			SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
		a.Pages.AddPage(config.SDIAG_PAGE, a.SchedView, true, false)
	}

	{ // Starting position
		a.ActivatePage(config.StartPane)
	}
}

func (a *App) RefreshClusterMetadata() {
	h, c, v, err := config.GetSchedulerInfoWithTimeout(config.RequestTimeout)
	a.App.QueueUpdateDraw(func() {
		if err != nil {
			a.MainFlex.SetTitle(" [red]DISCONNECTED FROM CLUSTER[white] ")
			return
		}
		a.MainFlex.SetTitle(fmt.Sprintf(
			" stui on [%s / %s / Slurm %s] ", c, h, v,
		))
	})
}

func (a *App) UpdateAppTitle() {
	a.MainFlex.SetTitle(fmt.Sprintf(
		" stui on [%s / %s / Slurm %s] ", config.ClusterName, config.SchedulerHostName, config.SchedulerSlurmVersion,
	))
}

// Starts periodic background processes to refresh data
func (a *App) StartRefresh() {
	// Fetch and setup partitions list - static
	a.PartitionsData = model.EmptyTableData() // Initialize
	a.PartitionsData = a.PartitionsProvider.Data()

	// First render of all views
	a.NodesView.Render()
	a.JobsView.Render()
	a.SacctView.Render()
	a.SacctMgrView.Render()
	{ // Render sdiag
		d := a.SdiagProvider.Data()
		a.SchedView.SetText(d.Data)
	}
	a.FirstRenderComplete = true

	// Other one-off actions that can only take place post first render
	a.setupPartitionSelectorOptions()
	a.NodesView.Table.ScrollToBeginning()
	a.JobsView.Table.ScrollToBeginning()
	if config.SacctEnabled {
		a.SacctMgrView.Table.ScrollToBeginning()
		a.SacctView.Table.ScrollToBeginning()
	}

	// Set periodic refreshes running. To make this very light on the scheduler, we:
	// 1. Do a full fetch of all sources at provider instantiation time, depending on configured quickstart/start-pane options.
	// 2. After that, only fetch data periodically for the active pane (this bit)
	// 3. On switching panes, if the data is older than refresh interval, we trigger a background refresh
	//    this happens in the key binds file. (keybinds.go)
	go func() {
		fetchTicker := time.NewTicker(config.RefreshInterval)
		defer fetchTicker.Stop()

		for {
			select {
			case <-fetchTicker.C:
				go a.RefreshClusterMetadata()
				a.App.QueueUpdateDraw(func() {
					switch a.GetCurrentPageName() {
					case config.NODES_PAGE:
						a.NodesView.FetchAndRender()
					case config.JOBS_PAGE:
						a.JobsView.FetchAndRender()
					case config.SACCTMGR_PAGE:
						a.SacctMgrView.FetchAndRender()
					case config.SACCT_PAGE:
						a.SacctView.FetchAndRender()
					case config.SDIAG_PAGE:
						a.SdiagProvider.Fetch()
						a.SchedView.SetText(a.SdiagProvider.Data().Data)
					}
				})
			}
		}
	}()
}

// Sets current view and updates inner grid contents for selectors
func (a *App) ActivatePage(page string) {
	switch page {
	case config.NODES_PAGE:
		{
			a.SwitchToPage(config.NODES_PAGE)
			a.CurrentTableView = a.NodesView.Table
			a.SetHeaderGridInnerContents(
				a.PartitionSelector,
				a.NodeStateSelector,
				a.SortSelector,
			)
			if a.SearchPattern != "" {
				a.ShowSearchBox(a.NodesView.Grid)
			} else {
				a.HideSearchBox()
			}
			a.App.SetFocus(a.NodesView.Table)
			a.setupSortSelectorOptions(a.NodesProvider, a.NodesView.sortColumn)
			a.PagesContainer.SetTitle(a.NodesView.completeTitle)
			go a.App.QueueUpdateDraw(func() {
				a.NodesView.FetchIfStaleAndRender(config.RefreshInterval)
			})
		}
	case config.JOBS_PAGE:
		{
			a.SwitchToPage(config.JOBS_PAGE)
			a.CurrentTableView = a.JobsView.Table
			a.SetHeaderGridInnerContents(
				a.PartitionSelector,
				a.JobStateSelector,
				a.SortSelector,
			)
			if a.SearchPattern != "" {
				a.ShowSearchBox(a.JobsView.Grid)
			} else {
				a.HideSearchBox()
			}
			a.App.SetFocus(a.JobsView.Table)
			a.setupSortSelectorOptions(a.JobsProvider, a.JobsView.sortColumn)
			a.PagesContainer.SetTitle(a.JobsView.completeTitle)
			go a.App.QueueUpdateDraw(func() {
				a.JobsView.FetchIfStaleAndRender(config.RefreshInterval)
			})
		}
	case config.SACCT_PAGE:
		{
			if config.SacctEnabled {
				a.SwitchToPage(config.SACCT_PAGE)
				a.CurrentTableView = a.SacctView.Table
				a.SetHeaderGridInnerContents(
					a.PartitionSelector,
					a.JobStateSelector,
					a.SortSelector,
				)
				if a.SearchPattern != "" {
					a.ShowSearchBox(a.SacctView.Grid)
				} else {
					a.HideSearchBox()
				}
				a.App.SetFocus(a.SacctView.Table)
				a.setupSortSelectorOptions(a.SacctProvider, a.SacctView.sortColumn)
				a.PagesContainer.SetTitle(a.SacctView.completeTitle)
				go a.App.QueueUpdateDraw(func() {
					a.SacctView.FetchIfStaleAndRender(config.RefreshInterval)
				})
			}
		}
	case config.SACCTMGR_PAGE:
		{
			if config.SacctEnabled {
				a.SwitchToPage(config.SACCTMGR_PAGE)
				a.CurrentTableView = a.SacctMgrView.Table
				a.SetHeaderGridInnerContents(
					a.SacctMgrEntitySelector,
					a.SortSelector,
				)
				if a.SearchPattern != "" {
					a.ShowSearchBox(a.SacctMgrView.Grid)
				} else {
					a.HideSearchBox()
				}
				a.App.SetFocus(a.SacctMgrView.Table)
				a.setupSortSelectorOptions(a.SacctMgrProvider, a.SacctMgrView.sortColumn)
				a.PagesContainer.SetTitle(a.SacctMgrView.completeTitle)
				go a.App.QueueUpdateDraw(func() {
					a.SacctMgrView.FetchIfStaleAndRender(config.RefreshInterval)
				})
			}
		}
	case config.SDIAG_PAGE:
		{
			a.SwitchToPage(config.SDIAG_PAGE)
			a.PagesContainer.SetTitle(" Scheduler status (sdiag) ")
			a.CurrentTableView = nil
			a.HideSearchBox()
			a.SetHeaderGridInnerContents(tview.NewBox())
			a.UpdateHeaderLineOne("")
			a.UpdateHeaderLineTwo("")
			a.App.SetFocus(a.SchedView)
			//a.SdiagProvider.Fetch()
			go a.App.QueueUpdateDraw(func() {
				a.SdiagProvider.FetchIfStale(config.RefreshInterval)
				a.SchedView.SetText(a.SdiagProvider.Data().Data) // This has no "render" function, we set it manually
			})
		}
	}
}
