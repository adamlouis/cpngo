let editorRef;
let cytoscapeRef;
let cytoscaleEl;

const layout = {
  name: "breadthfirst",
  directed: true,
};

const state = {
  didInit: false,
  updateEditor: true,
  useWasm: false,
  error: "",
  net: {
    places: [
      { id: "p1" },
      { id: "p2" },
      { id: "p3" },
      { id: "p4" },
      { id: "p5" },
    ],
    transitions: [{ id: "t1" }, { id: "t2" }, { id: "t3" }, { id: "t4" }],
    input_arcs: [
      { id: "p1t1", from_id: "p1", to_id: "t1" },
      { id: "p2t2", from_id: "p2", to_id: "t2" },
      { id: "p3t3", from_id: "p3", to_id: "t3" },
      { id: "p4t4", from_id: "p4", to_id: "t4" },
    ],
    output_arcs: [
      { id: "t1p2", from_id: "t1", to_id: "p2" },
      { id: "t1p3", from_id: "t1", to_id: "p3" },
      { id: "t2p4", from_id: "t2", to_id: "p4" },
      { id: "t3p4", from_id: "t3", to_id: "p4" },
      { id: "t4p5", from_id: "t4", to_id: "p5" },
      { id: "t4p1", from_id: "t4", to_id: "p1" },
    ],
    tokens: [{ id: "t1", place_id: "p1", color: "foobar" }],
  },
};

async function init() {
  if (state.didInit) {
    return;
  }
  state.didInit = true;

  cytoscaleEl = document.getElementById("net-cytoscape");

  editorRef = ace.edit("net-editor");
  editorRef.session.setMode("ace/mode/json");
  editorRef.getSession().on("change", function () {
    try {
      const net = JSON.parse(editorRef.getValue());
      state.net = net;
      render();
    } catch (e) {}
  });

  cytoscapeRef = cytoscape(createCyctoscapeData([]));

  render();

  const res = await WebAssembly.instantiateStreaming(
    fetch("cpngo.wasm"),
    go.importObject
  );
  state.useWasm = true;
  go.run(res.instance);
}

function createCyctoscapeData(elements) {
  return {
    container: cytoscaleEl,
    elements: elements,
    userPanningEnabled: false,
    userZoomingEnabled: false,
    style: [
      {
        selector: "node[kind='transition']",
        style: {
          "background-color": "#CE3262",
          shape: "rectangle",
        },
      },
      {
        selector: "node[kind='place']",
        style: {
          "background-color": "#00A29C",
          label: `data(tokens)`,
          "font-size": "36px",
        },
      },

      {
        selector: "edge",
        style: {
          width: 3,
          "line-color": "#000",
          "target-arrow-color": "#000",
          "target-arrow-shape": "triangle",
          "curve-style": "bezier",
        },
      },
    ],

    layout,
  };
}

function render() {
  document.getElementById("error-banner").innerText = state.error;

  if (state.updateEditor) {
    state.updateEditor = false;
    editorRef.session.setValue(JSON.stringify(state.net, null, 2));
  }

  const tokensByPlace = {};
  for (const token of state.net.tokens) {
    if (!tokensByPlace[token.place_id]) {
      tokensByPlace[token.place_id] = [];
    }
    tokensByPlace[token.place_id].push(token);
  }

  const elements = [];
  for (const place of state.net.places) {
    const count = `${tokensByPlace[place.id]?.length || 0}`;
    elements.push({
      data: {
        id: place.id,
        kind: "place",
        tokens: `${count}`,
      },
    });
  }
  for (const transition of state.net.transitions) {
    elements.push({
      data: {
        id: transition.id,
        kind: "transition",
      },
    });
  }
  for (const arc of state.net.input_arcs) {
    elements.push({
      data: {
        id: arc.id,
        source: arc.from_id,
        target: arc.to_id,
        kind: "input_arc",
      },
    });
  }
  for (const arc of state.net.output_arcs) {
    elements.push({
      data: {
        id: arc.id,
        source: arc.from_id,
        target: arc.to_id,
        kind: "output_arc",
      },
    });
  }

  cytoscapeRef.json(createCyctoscapeData(elements));
  cytoscapeRef.layout(layout).run();
}

function fire() {
  if (state.useWasm) {
    fireWasm();
  } else {
    fireServer();
  }
}

async function fireServer() {
  const res = await fetch("/fire", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ net: state.net }),
  });

  let txt = "";
  try {
    state.net = JSON.parse(await res.text()).net;
    state.updateEditor = true;
    state.error = "";
  } catch (err) {
    state.error = txt;
  }
  render();
}

const go = new Go();

async function fireWasm() {
  let txt = "";
  try {
    txt = GoFire(JSON.stringify(state.net));
    state.net = JSON.parse(txt);
    state.updateEditor = true;
    state.error = "";
  } catch (err) {
    state.error = `${txt}`;
  }
  render();
}
