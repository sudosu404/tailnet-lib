import { load } from "./node_modules/@starfederation/datastar/dist/bundles/datastar.js";
load();


window.sortList = function() {
  const list = document.getElementById("proxy-list");
  if (!list) return;

  const items = [...list.children].sort((a, b) => {
    return a.id.localeCompare(b.id);
  });

  items.forEach(item => list.appendChild(item));
}


