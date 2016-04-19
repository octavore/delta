/*eslint-env browser*/
/*global PouchDB:false */

const filesetExpiration = 10*60*1000; // filesets expire in 10 minutes.
const filesetGroupingWindow = 5000;
const pouch = new PouchDB("delta");

export function hasMore(dirhash, ts) {
  let start = `dm-${dirhash}-${ts}`;
  let end = `dm-${dirhash}-${ts+filesetGroupingWindow+1}-`;
  return pouch.allDocs({
    startkey: start,
    endkey: end,
  }).then((result) => {
    return result.rows.length > 1;
  }).catch((err) => {
    console.log("hasMore error:");
    console.log(err);
  });
}

export function collect(dirhash, ts, callback) {
  let start = `dm-${dirhash}-${ts-filesetGroupingWindow}`;
  let end = `dm-${dirhash}-${ts+filesetGroupingWindow+1}-`;
  return pouch.allDocs({
    startkey: start,
    endkey: end,
    include_docs: true
  }).then((result) => {
    return result.rows.map((r) => callback(r.doc));
  }).catch((err) => {
    console.log("collect error:");
    console.log(err);
  });
}

// storage: keyed on working dir plus date
// addFile is async because saving metadata may
// require retries.
export function addFile(metadata, file) {
  metadata._id = `dm-${metadata.dirhash}-${metadata.timestamp}-${metadata.hash}`;
  metadata.fileKey = metadata._id.replace("dm-", "df-");
  return pouch.put(metadata).then((response) => {
    return pouch.put({
      _id: metadata.fileKey,
      diff: file,
    });
  }).catch((err) => {
    console.log("addFile error:");
    console.log(err);
  });
}

// getFile returns a Promise
export function getFile(metadata) {
  return pouch.get(metadata.fileKey).catch((err) => {
    console.log("getFile error:");
    console.log(err);
  });
}

// purge old documents on window unload.
window.addEventListener("load", () => {
  let d = new Date() * 1 - filesetExpiration;
  pouch.allDocs({
    startkey: "dm-",
    endkey: "dm-\uffff"
  }).then((result) => {
    result.rows.map((r) => {
      if (r.id.split("-")[2]*1 < d) {
        pouch.remove(r.id, r.value.rev);
      }
    });
  });

  pouch.allDocs({
    startkey: "df-",
    endkey: "df-\uffff"
  }).then((result) => {
    result.rows.map((r) => {
      if (r.id.split("-")[2]*1 < d) {
        pouch.remove(r.id, r.value.rev);
      }
    });
  });
});
