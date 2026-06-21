package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

type apiItem struct {
	Stem     string   `json:"stem"`
	Category string   `json:"category"`
	Models   []string `json:"models"`
}

type apiCategoryStats struct {
	Name    string         `json:"name"`
	Total   int            `json:"total"`
	ByModel map[string]int `json:"byModel"`
}

type apiStatus struct {
	Total      int                `json:"total"`
	Models     []string           `json:"models"`
	Categories []apiCategoryStats `json:"categories"`
	Items      []apiItem          `json:"items"`
}

func buildAPIStatus(dir string) (apiStatus, error) {
	items, allModels, err := scanStatus(dir)
	if err != nil {
		return apiStatus{}, err
	}

	cats := make(map[string]*apiCategoryStats)
	var catOrder []string

	var apiItems []apiItem
	for _, it := range items {
		apiItems = append(apiItems, apiItem{
			Stem:     it.stem,
			Category: it.category,
			Models:   it.models,
		})
		cs, ok := cats[it.category]
		if !ok {
			cs = &apiCategoryStats{Name: it.category, ByModel: make(map[string]int)}
			cats[it.category] = cs
			catOrder = append(catOrder, it.category)
		}
		cs.Total++
		for _, m := range it.models {
			cs.ByModel[m]++
		}
	}
	sort.Strings(catOrder)

	var catList []apiCategoryStats
	for _, c := range catOrder {
		catList = append(catList, *cats[c])
	}

	return apiStatus{
		Total:      len(items),
		Models:     allModels,
		Categories: catList,
		Items:      apiItems,
	}, nil
}

func serveDashboard(dir string, port int) error {
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status, err := buildAPIStatus(dir)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(status)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, dashboardHTML)
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("vtidy dashboard → http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="zh-Hant">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>vtidy dashboard</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; color: #1a1a1a; padding: 2rem; }
@media (prefers-color-scheme: dark) {
  body { background: #1a1a1a; color: #e5e5e5; }
  .card { background: #252525; border-color: #333; }
  .metric { background: #2a2a2a; }
  .table-header { background: #2a2a2a; }
  .table-row { border-color: #333; }
  .table-row:hover { background: #2a2a2a; }
  .badge-done { background: #0a3d2a; color: #4ade80; }
  .badge-partial { background: #3d2e0a; color: #fbbf24; }
  .badge-none { background: #3d0a0a; color: #f87171; }
  .bar-bg { background: #333; }
  .activity-item { border-color: #333; }
  .refresh-btn { background: #333; color: #e5e5e5; border-color: #444; }
  .refresh-btn:hover { background: #444; }
}
.container { max-width: 800px; margin: 0 auto; }
.header { display: flex; align-items: center; gap: 10px; margin-bottom: 1.5rem; }
.header h1 { font-size: 20px; font-weight: 500; }
.header .dir { font-size: 13px; color: #888; margin-left: auto; }
.metrics { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 1.5rem; }
.metric { background: #fff; border-radius: 10px; padding: 1rem; }
.metric .label { font-size: 13px; color: #888; }
.metric .value { font-size: 28px; font-weight: 500; margin-top: 2px; }
.metric .value.green { color: #16a34a; }
.metric .value.amber { color: #d97706; }
.card { background: #fff; border: 0.5px solid #e5e5e5; border-radius: 12px; overflow: hidden; margin-bottom: 1.5rem; }
.card-title { font-size: 14px; font-weight: 500; padding: 12px 16px; border-bottom: 0.5px solid #e5e5e5; }
.table-header { display: grid; padding: 8px 16px; font-size: 12px; color: #888; background: #fafafa; }
.table-row { display: grid; padding: 12px 16px; font-size: 14px; align-items: center; border-bottom: 0.5px solid #f0f0f0; }
.table-row:last-child { border-bottom: none; }
.table-row:hover { background: #fafafa; }
.table-row .cat { font-weight: 500; }
.bar-bg { height: 6px; background: #f0f0f0; border-radius: 3px; overflow: hidden; }
.bar-fill { height: 100%; border-radius: 3px; transition: width 0.3s; }
.bar-fill.green { background: #16a34a; }
.bar-fill.amber { background: #d97706; }
.bar-fill.red { background: #ef4444; }
.bar-label { font-size: 11px; color: #888; margin-top: 2px; }
.badge { font-size: 12px; padding: 2px 10px; border-radius: 6px; display: inline-block; text-align: center; min-width: 48px; }
.badge-done { background: #dcfce7; color: #166534; }
.badge-partial { background: #fef3c7; color: #92400e; }
.badge-none { background: #fee2e2; color: #991b1b; }
.activity-item { display: flex; align-items: center; gap: 10px; padding: 10px 16px; border-bottom: 0.5px solid #f0f0f0; font-size: 13px; }
.activity-item:last-child { border-bottom: none; }
.dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.dot.green { background: #16a34a; }
.dot.amber { background: #d97706; }
.dot.red { background: #ef4444; }
.path-dim { color: #888; }
.refresh-btn { background: #fff; border: 0.5px solid #e5e5e5; border-radius: 8px; padding: 8px 16px; font-size: 13px; cursor: pointer; display: flex; align-items: center; gap: 6px; }
.refresh-btn:hover { background: #fafafa; }
.spin { animation: spin 0.8s linear infinite; display: inline-block; }
@keyframes spin { to { transform: rotate(360deg); } }
.footer { display: flex; align-items: center; gap: 12px; }
.auto-label { font-size: 12px; color: #888; }
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <h1>vtidy dashboard</h1>
    <span class="dir" id="dir">resources/</span>
  </div>
  <div class="metrics" id="metrics"></div>
  <div class="card">
    <div class="card-title">By category</div>
    <div id="category-header"></div>
    <div id="category-rows"></div>
  </div>
  <div class="card">
    <div class="card-title">All transcripts</div>
    <div id="item-header"></div>
    <div id="item-rows"></div>
  </div>
  <div class="footer">
    <button class="refresh-btn" id="refresh-btn" onclick="loadData()">&#x21bb; Refresh</button>
    <span class="auto-label" id="auto-label"></span>
  </div>
</div>
<script>
let data = null;

async function loadData() {
  const btn = document.getElementById('refresh-btn');
  btn.innerHTML = '<span class="spin">&#x21bb;</span> Loading...';
  try {
    const res = await fetch('/api/status');
    data = await res.json();
    render();
  } catch(e) {
    console.error(e);
  }
  btn.innerHTML = '&#x21bb; Refresh';
  document.getElementById('auto-label').textContent = 'Updated ' + new Date().toLocaleTimeString();
}

function badge(count, total) {
  if (count === total) return '<span class="badge badge-done">' + count + '/' + total + '</span>';
  if (count > 0) return '<span class="badge badge-partial">' + count + '/' + total + '</span>';
  return '<span class="badge badge-none">0/' + total + '</span>';
}

function barClass(count, total) {
  if (count === total) return 'green';
  if (count > 0) return 'amber';
  return 'red';
}

function render() {
  if (!data) return;

  let totalDone = 0;
  let totalPending = 0;
  const modelCount = data.models.length;

  data.items.forEach(it => {
    if (it.models.length === modelCount) totalDone++;
    else totalPending += modelCount - it.models.length;
  });

  document.getElementById('metrics').innerHTML =
    metric('Total transcripts', data.total, '') +
    metric('Fully processed', totalDone, 'green') +
    metric('Missing outputs', totalPending, totalPending > 0 ? 'amber' : 'green');

  const cols = '140px 1fr' + data.models.map(() => ' 90px').join('');

  let hdr = '<div class="table-header" style="grid-template-columns:' + cols + '"><span>Category</span><span>Progress</span>';
  data.models.forEach(m => { hdr += '<span style="text-align:center">' + m + '</span>'; });
  hdr += '</div>';
  document.getElementById('category-header').innerHTML = hdr;

  let rows = '';
  data.categories.forEach(cat => {
    let processed = 0;
    data.models.forEach(m => { processed += (cat.byModel[m] || 0); });
    const maxProcessed = cat.total * modelCount;
    const pct = maxProcessed > 0 ? Math.round(processed / maxProcessed * 100) : 0;

    rows += '<div class="table-row" style="grid-template-columns:' + cols + '">';
    rows += '<span class="cat">' + cat.name + '</span>';
    rows += '<div><div class="bar-bg"><div class="bar-fill ' + barClass(pct, 100) + '" style="width:' + pct + '%"></div></div>';
    rows += '<div class="bar-label">' + processed + '/' + maxProcessed + ' outputs</div></div>';
    data.models.forEach(m => {
      rows += '<span style="text-align:center">' + badge(cat.byModel[m] || 0, cat.total) + '</span>';
    });
    rows += '</div>';
  });
  document.getElementById('category-rows').innerHTML = rows;

  let ihdr = '<div class="table-header" style="grid-template-columns:140px 1fr' + data.models.map(() => ' 60px').join('') + '"><span>Category</span><span>Transcript</span>';
  data.models.forEach(m => { ihdr += '<span style="text-align:center">' + shortModel(m) + '</span>'; });
  ihdr += '</div>';
  document.getElementById('item-header').innerHTML = ihdr;

  let irows = '';
  data.items.forEach(it => {
    irows += '<div class="table-row" style="grid-template-columns:140px 1fr' + data.models.map(() => ' 60px').join('') + '">';
    irows += '<span class="path-dim">' + it.category + '</span>';
    irows += '<span>' + it.stem + '</span>';
    data.models.forEach(m => {
      const has = it.models.includes(m);
      irows += '<span style="text-align:center"><span class="dot ' + (has ? 'green' : 'red') + '" title="' + m + '"></span></span>';
    });
    irows += '</div>';
  });
  document.getElementById('item-rows').innerHTML = irows;
}

function metric(label, value, cls) {
  return '<div class="metric"><div class="label">' + label + '</div><div class="value ' + cls + '">' + value + '</div></div>';
}

function shortModel(m) {
  if (m.length > 8) return m.substring(0, 7) + '…';
  return m;
}

loadData();
setInterval(loadData, 10000);
</script>
</body>
</html>`
