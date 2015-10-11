/*eslint-env browser*/
/*global PouchDB:false */

const filesetExpiration = 5*60000; // filesets expire in 5 minutes.
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
