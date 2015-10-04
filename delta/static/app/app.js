/*eslint-env browser*/
/*global m:false hljs:false metadata:false */

import path from "path";
import * as storage from "./lib/storage";
import langMap from "./lib/lang";

class AppController {
  constructor() {
    // fileList contains a list of Entries
    metadata.data = document.querySelector("#diff").innerHTML;
    this.currentFile = m.prop();
    this.setCurrentFile(metadata);
    storage.addFile(metadata);

    this.fileList = m.prop([]);
    this.showMenu = m.prop(true);
    this.showContext = m.prop(true);
    this.showEmpty = m.prop(true);

    // polling to detect if storage changes. if it changes,
    // then close this tab because that indicates another was
    // opened. only poll for 1 second, after that consider this tab
    // permanent.
    this.maxDelayMillis = 3000;
    this.size = storage.size(metadata.dir);
    this.start = new Date();
    this.poll = setInterval(this._poll.bind(this), 20);
  }

  _poll() {
    let now = new Date();
    if ((now - this.start) > this.maxDelayMillis) {
      clearInterval(this.poll);
    }
    if (storage.size(metadata.dir) != this.size) {
      window.close();
      // todo: save a copy of the storage so refresh works?
    }
  }

  setCurrentFile(f) {
    document.title = f.merged;
    this.currentFile(f);
  }
}

// sidebarEntry creates the div for each sidebar entry.
// @param {CompareController} ctrl
// @param {FilePair} el
function sidebarEntry(ctrl, el) {
  let k = `.sidebar-entry-${el.change}`;
  if (el.merged === ctrl.currentFile().merged) {
    k += ".sidebar-entry-selected";
  }
  return m(".sidebar-entry" + k, {
    onclick: () => ctrl.setCurrentFile(el)
  }, path.basename(el.merged));
}

window.App = {
  controller: AppController,
  view: (ctrl) => {
    let doc = document.createElement("div");
    if (ctrl.currentFile() != null) {
      doc.innerHTML = ctrl.currentFile().data;
      let diffs = doc.querySelectorAll(".diff-pane");
      for (let i = 0; i < diffs.length; i++) {
        let lang = langMap[path.extname(ctrl.currentFile().merged)];
        if (lang != null) {
          diffs[i].classList.add(`lang-${lang}`);
        }
        hljs.highlightBlock(diffs[i]);

        // remove extra nodes that hljs adds sometimes
        let l = diffs[i].querySelectorAll(".diff-pane-contents > :not(.line)");
        for (var j = 0; j < l.length; j++) { l[j].remove(); }
      }
    }

    return m("#contents", [
      m(`#sidebar.sidebar-show-${ctrl.showMenu()}`,
        m(".sidebar-inner",
          m("a.sidebar-header", {href: "https://github.com/octavore/delta"}, "Delta"),
          storage.collect(metadata.dir, (key, data) => sidebarEntry(ctrl, data))
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
