function toggleNavigationSidebar() {
  const menu = document.getElementById("navigation")
  const backdrop = document.getElementById("backdrop")
  menu.classList.toggle("show")
  backdrop.classList.toggle("show")
}

async function copyLinkToClipboard() {
  await navigator.clipboard.writeText(window.location.href);
}

function searchArticles(e) {
  const query = e.value.trim();
  let newURL = window.location.pathname;

  if ("" !== query) {
    newURL += "?search=" + query;
  }

  e.setAttribute("hx-get", newURL);
  e.setAttribute("hx-push-url", newURL);
  window.history.replaceState({}, "", newURL);
}

function setArchiveTopic(e) {
  document.querySelector("h3.topic-and-date").classList.remove("hide");
  document.querySelector("div.selected-tag-div").classList.add("hide");
  document.querySelector("span.selected-topic").textContent = e.textContent;
  document.querySelector("span.selected-date").textContent = "Any date";

  for (const topicItem of document.getElementById("topics-list").children) {
    topicItem.classList.remove("selected");
  }

  for (const tagItem of document.getElementById("tags-list").children) {
    tagItem.classList.remove("selected");
  }

  const currentTopicItem = e.parentNode;
  currentTopicItem.classList.add("selected");

  const baseTopicURL = e.getAttribute("hx-get");
  const publicationsList = document.getElementById("publications-list").children;

  for (const publicationItem of publicationsList) {
    publicationItem.classList.remove("selected");

    let newTopicURL = baseTopicURL;
    const anchor = publicationItem.firstChild;
    const hxGetValue = anchor.getAttribute("hx-get").split("/");
    const month = hxGetValue.at(-1);
    const year = hxGetValue.at(-2);

    newTopicURL += "/" + year + "/" + month;
    anchor.setAttribute("href", newTopicURL);
    anchor.setAttribute("hx-get", newTopicURL);
    anchor.setAttribute("hx-push-url", "true");
  }

  const searchbar = document.getElementById("searchbar");

  searchbar.value = "";

  searchbar.setAttribute("hx-get", baseTopicURL);
  searchbar.setAttribute("hx-push-url", baseTopicURL);
  window.history.replaceState({}, "", baseTopicURL);

  htmx.process(document.body);
}

function setArchivePublicationDate(e) {
  document.querySelector("h3.topic-and-date").classList.remove("hide");
  document.querySelector("div.selected-tag-div").classList.add("hide");
  document.querySelector("span.selected-date").textContent = e.textContent;

  for (const publicationItem of document.getElementById("publications-list").children) {
    publicationItem.classList.remove("selected");
  }

  for (const tagItem of document.getElementById("tags-list").children) {
    tagItem.classList.remove("selected");
  }

  const currentPublicationItem = e.parentNode;
  const currentURL = e.getAttribute("href");

  currentPublicationItem.classList.add("selected");

  const searchbar = document.getElementById("searchbar");

  searchbar.value = "";
  searchbar.setAttribute("hx-get", currentURL);
  searchbar.setAttribute("hx-push-url", currentURL);

  window.history.replaceState({}, "", currentURL);

  htmx.process(document.body);
}

function setArchiveTag(e) {
  document.querySelector("h3.topic-and-date").classList.add("hide");
  document.querySelector("div.selected-tag-div").classList.remove("hide");
  document.querySelector(".selected-tag").textContent = e.textContent;

  for (const topicItem of document.getElementById("topics-list").children) {
    topicItem.classList.remove("selected");
  }

  for (const tagItem of document.getElementById("tags-list").children) {
    tagItem.classList.remove("selected");
  }

  const currentTagItem = e.parentNode;
  currentTagItem.classList.add("selected");

  const baseTagURL = e.getAttribute("hx-get");

  for (const publicationItem of document.getElementById("publications-list").children) {
    publicationItem.classList.remove("selected");

    let publicationURL = "/archive/any";
    const anchor = publicationItem.firstChild;
    const hxGetValue = anchor.getAttribute("hx-get").split("/");
    const month = hxGetValue.at(-1);
    const year = hxGetValue.at(-2);

    publicationURL += "/" + year + "/" + month;
    anchor.setAttribute("href", publicationURL);
    anchor.setAttribute("hx-get", publicationURL);
    anchor.setAttribute("hx-push-url", "true");
  }

  const searchbar = document.getElementById("searchbar");

  searchbar.value = "";

  searchbar.setAttribute("hx-get", baseTagURL);
  searchbar.setAttribute("hx-push-url", baseTagURL);
  window.history.replaceState({}, "", baseTagURL);

  htmx.process(document.body);
}
