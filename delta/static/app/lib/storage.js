/*eslint-env browser*/

/*

data is stored as follows:

    localStorage[deltaKey] = {
      "metadata.dir": {
        "_time": new Date() * 1;
        "metadata.merged": {}
      }
    }

*/

const deltaKey = "delta";

function getGlobalStore() {
  let globalStore = localStorage.getItem(deltaKey);
  if (typeof globalStore == "string") {
    globalStore = JSON.parse(globalStore);
  }
  if (globalStore == null || typeof globalStore != "object") {
    globalStore = {};
    localStorage.setItem(deltaKey, JSON.stringify(globalStore));
  }
  return globalStore;

}

function getStore(workingDir) {
  let globalStore = getGlobalStore();
  if (!(workingDir in globalStore)) {
    localStorage.setItem(deltaKey, JSON.stringify(globalStore));
  }
  return globalStore[workingDir];
}

function updateStore(workingDir, filePath, fileMetadata) {
  let now = new Date();
  let globalStore = getGlobalStore();
  if (!(workingDir in globalStore) || (now - globalStore[workingDir]["_created_at"]) > 1000) {
    globalStore[workingDir] = {"_created_at": now*1};
  }
  globalStore[workingDir][filePath] = fileMetadata;
  localStorage.setItem(deltaKey, JSON.stringify(globalStore));
}

export function size(dir) {
  return Object.keys(getGlobalStore()[dir]).length;
}

export function collect(dir, callback) {
  let out = [];
  let storage = getStore(dir);
  for (var key in storage) {
    if (key == "_created_at") {
      continue;
    }
    out.push(callback(key, storage[key]));
  }
  return out;
}

// storage: keyed on working dir plus date
export function addFile(metadata) {
  // todo: catch errors
  updateStore(metadata.dir, metadata.merged, metadata);
}
