/*eslint-env browser*/
/*global m:false hljs:false Mousetrap:false metadata:false */

import path from "path";
import * as storage from "./lib/storage";
import langMap from "./lib/lang";

let defaultConfig = {
  // context is the number of lines of surrounding lines to show.
  // valid values are 0-4, and `"false"` to show all context
  context: 4,

  // showEmpty toggles hiding of empty lines in the diff. Empty diff lines are
  // created when there is an added or deleted line.
  showEmpty: true,

  // shouldCollapse toggles collapsing of browser tabs for a group of diffs
  // into a single tab.
  shouldCollapse: true,

  // highlight toggles syntax highlighting
  highlight: true,

  // unmodifiedOpacity controls the opacity of unmodified lines. Set to
  // 0 to use the stylesheet default
  unmodifiedOpacity: false,

  // diffFontSize controls the font size in the diff. Set to 0 to use the
  // stylesheet default
  diffFontSize: false,
};

function merge(a, b={}) {
  let out = {};
  Object.keys(a).forEach((k) => out[k] = a[k]);
  Object.keys(b).forEach((k) => out[k] = b[k]);
  return out;
}

class AppController {
  constructor(config) {
    this.currentFile = m.prop(metadata);
    this.currentDiff = m.prop(document.querySelector("#diff").innerHTML);
    this.fileSaved = false;
    storage.addFile(metadata, this.currentDiff()).then(() => {
      this.setCurrentFile(metadata);
      this.fileSaved = true;
    });

    this.config = merge(defaultConfig, config);
    this.fileGroups = m.prop([]);
    this.fileList = m.prop([]);
    this.showMenu = m.prop(false);
    this.showContext = m.prop(this.config.context+1);
    this.showEmpty = m.prop(this.config.showEmpty);
    // polling to detect if storage changes. if it changes,
    // then close this tab because that indicates another was
    // opened. only poll for 1 second, after that consider this tab
    // permanent.
    // todo: move this to the storage via the ability to register callbacks?
    this.maxDelayMillis = 2000;
    this.start = new Date();
    this.poll = setInterval(this._poll.bind(this), 10);

    this._initKeyBindings();
  }

  _poll() {
    // abort if the file has not been saved or collapsing has been disabled
    if (!this.fileSaved || !this.config.shouldCollapse) {
      return;
    }

    this.updateSidebar();
    let now = new Date();

    // if there is no change to the change list for this dir within
    // maxDelayMillis, then clear the poll.
    if ((now - this.start) > this.maxDelayMillis) {
      clearInterval(this.poll);
    }

    // another tab opened a diff for the same dir (within maxDelayMillis)
    // so close this tab.
    // todo: save a copy of the storage so refresh works?
    storage.hasMore(metadata.dirhash, metadata.timestamp).then((hasMore) => {
      if (hasMore) window.close();
    });
  }

  _initKeyBindings() {
    for (var i = 0; i < 6; i++) {
      let j = i;
      Mousetrap.bind(`${i}`, () => {
        this.showContext(j);
        m.redraw();
      });
    };
    Mousetrap.bind("6", () => {
      this.showContext(10);
      m.redraw();
    });
    Mousetrap.bind("d", () => {
      this.showMenu(!this.showMenu());
      m.redraw();
    });
    Mousetrap.bind("j", this.nextFile.bind(this));
    Mousetrap.bind("k", this.prevFile.bind(this));
  }

  setCurrentFile(metadata) {
    document.title = metadata.merged;
    storage.getFile(metadata).then((blob) => {
      this.currentFile(metadata);
      this.currentDiff(blob.diff);
      this.updateSidebar();
      m.redraw();
    });
  }

  updateSidebar() {
    let groups = {};
    let fileList = [];
    this.fileList([]);
    if (!this.config.shouldCollapse) {
      return;
    }
    storage.collect(metadata.dirhash, metadata.timestamp, (meta) => {
      let dir = path.dirname(meta.merged);
      groups[dir] = groups[dir] || [];
      groups[dir].push(meta);
    }).then(() => {
      let sortedGroups = Object.keys(groups).sort();
      this.fileGroups(sortedGroups.map((group) => {
        fileList = fileList.concat(groups[group]);
        return {dir: group, files: groups[group]};
      }));
      if (fileList.length > 1) {
        this.showMenu(true);
      }
      this.fileList(fileList);
      m.redraw();
    });
  }

  nextFile() {
    let id = this.currentFile()._id;
    this.fileList().map((meta, i, lst) => {
      if (meta._id != id) {
        return;
      }
      if (lst.length > i+1) {
        this.setCurrentFile(lst[i+1]);
      }
    });
  }

  prevFile() {
    let id = this.currentFile()._id;
    this.fileList().map((meta, i, lst) => {
      if (meta._id != id) {
        return;
      }
      if (i > 0) {
        this.setCurrentFile(lst[i-1]);
      }
    });
  }
}

function sidebar(dir, ctrl) {
  return ctrl.fileGroups().map((group) => {
    let h = m(".sidebar-subheader", group.dir);
    if (group.dir == ".") {
      h = m(".sidebar-subheader", "<root>");
    }
    return m("div", h, group.files.map((meta) => {
      let k = `.sidebar-entry-${meta.change}`;
      if (meta.merged === ctrl.currentFile().merged) {
        k += ".sidebar-entry-selected";
      }
      return m(".sidebar-entry" + k, {
        onclick: () => ctrl.setCurrentFile(meta)
      }, path.basename(meta.merged));
    }));
  });
}

window.App = (config) => {
  return {
    controller: () => new AppController(config),
    view: (ctrl) => {
      let doc = document.createElement("div");
      if (ctrl.currentFile() != null) {
        doc.innerHTML = ctrl.currentDiff();
        let diffs = doc.querySelectorAll(".diff-pane");
        if (ctrl.config.highlight) {
          for (let i = 0; i < diffs.length; i++) {
            let lang = langMap[path.extname(ctrl.currentFile().merged)];
            if (lang != null) {
              diffs[i].classList.add(`lang-${lang}`);
            }

            // highlighting long things takes a while.
            if (diffs[i].innerHTML.length < 1000000) {
              hljs.highlightBlock(diffs[i]);
            }

            // remove extra nodes that hljs adds sometimes
            let l = diffs[i].querySelectorAll(".diff-pane-contents > :not(.line)");
            for (var j = 0; j < l.length; j++) { l[j].remove(); }
          }
        }
      }

      let style = "";
      if (parseFloat(ctrl.config.unmodifiedOpacity) > 0) {
        style += `.lm { opacity: ${parseFloat(ctrl.config.unmodifiedOpacity)} !important} `;
      }
      if (parseFloat(ctrl.config.diffFontSize) > 0) {
        style += `.diff-row { font-size: ${parseFloat(ctrl.config.diffFontSize)}px !important} `;
      }

      return m("#contents", [
        m("style", style),
        m(`#sidebar.sidebar-show-${ctrl.showMenu()}`,
          m(".sidebar-inner",
            m("a.sidebar-header", {href: "https://github.com/octavore/delta"}, "Delta"),
            sidebar(metadata.dir, ctrl)
          )
        ),
        m("#diff", ctrl.currentFile() == null ? null :
          [
            m(".diff-row.diff-row-headers",
              ctrl.currentFile().merged != null ?
                m(".diff-pane", ctrl.currentFile().merged) : [
                m(".diff-pane", ctrl.currentFile().from),
                m(".diff-pane", ctrl.currentFile().to)
              ]
            ),
            m(".diff-row-padding"),
            m(`.diff-row.diff-context-${ctrl.showContext()}.diff-empty-${ctrl.showEmpty()}`, {
              config: (el) => {
                el.innerHTML = "";
                while (doc.childNodes.length > 0) {
                  el.appendChild(doc.childNodes[0]);
                }
              }
            })
          ])
      ]);
    }
  };
};
