package view

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	NODES_PAGE    = "nodes"
	JOBS_PAGE     = "jobs"
	SACCTMGR_PAGE = "sacctmgr"
	SDIAG_PAGE    = "sdiag"
	COMMAND_PAGE  = "commandmodal"
)

type App struct {
	App                 *tview.Application
	Pages               *tview.Pages
	PagesContainer      *tview.Flex  // Container for pages with border title
	startTime           time.Time    // Start time of the application
	CurrentTableView    *tview.Table // Points to either NodesView or JobsView
	FirstRenderComplete bool

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
	TabNodesBox      *tview.TextView
	TabJobsBox       *tview.TextView
	TabSchedulerBox  *tview.TextView
	TabAccountingBox *tview.TextView

	// Dropdown selectors
	PartitionSelector      *tview.DropDown
	SacctMgrEntitySelector *tview.DropDown
	NodeStateSelector      *tview.DropDown
	JobStateSelector       *tview.DropDown

	// Search state
	SearchBox     *tview.InputField
	SearchActive  bool
	SearchPattern string
	searchTimer   *time.Timer // Timer for debouncing search updates

	// Command modal state
	CommandModalOpen bool

	// Stored Data
	JobsTableData  *model.TableData
	AcctTableData  *model.TableData
	PartitionsData *model.TableData

	// Data  and providers
	SchedulerHostName     string
	SchedulerClusterName  string
	SchedulerSlurmVersion string
	PartitionsProvider    model.DataProvider[*model.TableData]
	NodesProvider         model.DataProvider[*model.TableData]
	JobsProvider          model.DataProvider[*model.TableData]
	SacctMgrProvider      model.DataProvider[*model.TableData]
	SdiagProvider         model.DataProvider[*model.TextData]

	// New style views
	NodesView    *StuiView
	JobsView     *StuiView
	SacctMgrView *StuiView
	SchedView    *tview.TextView // Special case, text only
}

// Exit and log error details
func (a *App) closeOnError(err error) {
	if err != nil {
		a.App.Stop()
		log.Fatal(err)
	}
}

// Initializes a `stui` instance tview Application using the config module
func InitializeApplication() *App {
	application := App{
		startTime:               time.Now(),
		App:                     tview.NewApplication(),
		Pages:                   tview.NewPages(),
		HeaderGridInnerContents: tview.NewGrid(),
		FirstRenderComplete:     false,
	}

	// Init data providers at start - in parallel, as they all do their first fetch on initialization
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(6)
	go func() {
		defer wg.Done()
		application.PartitionsProvider = model.NewPartitionsProvider()
	}()
	go func() {
		defer wg.Done()
		application.NodesProvider = model.NewNodesProvider()
	}()
	go func() {
		defer wg.Done()
		application.JobsProvider = model.NewJobsProvider()
	}()
	go func() {
		defer wg.Done()
		application.SdiagProvider = model.NewSdiagProvider()
	}()
	go func() {
		defer wg.Done()
		application.SchedulerHostName, application.SchedulerClusterName, application.SchedulerSlurmVersion = model.GetSchedulerInfoWithTimeout(config.RequestTimeout)
	}()
	go func() {
		defer wg.Done()
		if config.SacctEnabled {
			application.SacctMgrProvider = model.NewSacctMgrProvider()
		}
	}()
	wg.Wait()
	if !config.Quiet {
		log.Printf("START: Initial data load from scheduler took %d ms", time.Since(start).Milliseconds())
	}
	return &application
}

func (a *App) SetupViews() {
	a.SetupSearchBox()
	a.SetupPartitionSelector()
	a.SetupNodeStateSelector()
	a.SetupJobStateSelector()

	if config.SacctEnabled {
		a.SetupSacctMgrEntitySelector()
	}

	{ // Headerlines
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
			SetText("(1) Nodes")
		a.TabNodesBox.SetBackgroundColor(paneSelectorHighlightColor)
		a.TabJobsBox = tview.NewTextView().
			SetText("(2) Jobs")
		a.TabSchedulerBox = tview.NewTextView().
			SetText("(3) Scheduler")
		a.TabAccountingBox = tview.NewTextView().
			SetText("(4) Acct Manager")
	}

	// Create a grid for the tabs
	tabGrid := tview.NewGrid().
		AddItem(a.TabNodesBox, FRST_ROW, FRST_COL, 1, 1, 1, 0, false).
		AddItem(a.TabJobsBox, SCND_ROW, FRST_COL, 1, 1, 1, 0, false).
		AddItem(a.TabSchedulerBox, THRD_ROW, FRST_COL, 1, 1, 1, 0, false)

	if config.SacctEnabled {
		tabGrid.AddItem(a.TabAccountingBox, FRTH_ROW, FRST_COL, 1, 1, 1, 0, false)
	}

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
		AddItem(a.HeaderGrid, 6, 0, false).
		AddItem(a.PagesContainer, 0, 1, true)

	a.MainFlex.SetBorder(true).
		SetBorderAttributes(tcell.AttrDim).
		SetTitle(fmt.Sprintf(
			" stui on [%s / %s / Slurm %s] ", a.SchedulerClusterName, a.SchedulerHostName, a.SchedulerSlurmVersion,
		)).
		SetTitleAlign(tview.AlignCenter)

	{ // Nodes View
		a.NodesView = NewStuiView(
			"Nodes",
			a.NodesProvider,
			a.PagesContainer.SetTitle,
			a.UpdateHeaderLineTwo, // errors
			a.UpdateHeaderLineOne, // data updates notify
		)
		a.Pages.AddPage(NODES_PAGE, a.NodesView.Grid, true, true)
	}

	{ // Jobs View
		a.JobsView = NewStuiView(
			"Jobs",
			a.JobsProvider,
			a.PagesContainer.SetTitle,
			a.UpdateHeaderLineTwo, // errors
			a.UpdateHeaderLineOne, // data updates notify
		)
		a.Pages.AddPage(JOBS_PAGE, a.JobsView.Grid, true, false)
	}

	{
		// Accounting view - we create this view whether not it will be used.
		// This way we do not need to gate our code everywhere to check for
		// whether it's enabled, just to avoid segfaults.
		a.SacctMgrView = NewStuiView(
			model.SACCTMGR_TABLE_ENTITIES[0], // First type of entity to start with
			a.SacctMgrProvider,
			a.PagesContainer.SetTitle,
			a.UpdateHeaderLineTwo, // errors
			a.UpdateHeaderLineOne, // data updates notify
		)

		a.Pages.AddPage(SACCTMGR_PAGE, a.SacctMgrView.Grid, true, false)
	}

	{ // Scheduler View
		a.SchedView = tview.NewTextView()
		a.SchedView.
			SetDynamicColors(true).
			SetScrollable(true).
			SetWrap(false).
			SetTitleAlign(tview.AlignLeft).
			SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
		a.Pages.AddPage(SDIAG_PAGE, a.SchedView, true, false)
	}

	{ // Starting position
		a.CurrentTableView = a.NodesView.Table
		a.SetHeaderGridInnerContents(
			a.PartitionSelector,
			a.NodeStateSelector,
		)
	}
}

// Starts periodic background processes to refresh data
func (a *App) StartRefresh() {
	// Fetch and setup partitions list - static
	a.PartitionsData = a.PartitionsProvider.Data()

	// First render of all views
	a.NodesView.Render()
	a.JobsView.Render()
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
	}

	// Set periodic refreshes running. To make this very light on the scheduler, we:
	// 1. Do a full fetch of all sources once, at the start
	// 2. After that, only fetch data periodically for the active pane
	// 3. On switching panes, if the data is older than refresh interval, we trigger a background refresh
	//    this happens in keybinds.
	go func() {
		renderTicker := time.NewTicker(3 * time.Second) // Render every 3 seconds, regardless of data refresh frequency
		fetchTicker := time.NewTicker(config.RefreshInterval)
		defer renderTicker.Stop()
		defer fetchTicker.Stop()

		for {
			select {
			case <-renderTicker.C:
				a.App.QueueUpdateDraw(func() {
					a.RenderCurrentView()
				})
			case <-fetchTicker.C:
				a.App.QueueUpdateDraw(func() {
					switch a.GetCurrentPageName() {
					case NODES_PAGE:
						a.NodesView.FetchAndRender()
					case JOBS_PAGE:
						a.JobsView.FetchAndRender()
					case SACCTMGR_PAGE:
						a.SacctMgrView.FetchAndRender()
					case SDIAG_PAGE:
						a.SdiagProvider.Fetch()
						a.SchedView.SetText(a.SdiagProvider.Data().Data)
					}
				})
			}
		}
	}()
}
