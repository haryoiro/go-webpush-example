self.addEventListener("push", function (event) {
  const data = event.data.json();
  const icon = data.icon;
  const title = data.title;
  const message = data.message;

  // safariでiconを通知するための処理
  const iconUrl = new URL(icon, location).href;
  console.log(iconUrl);

  const options = {
    body: message,
    icon: icon,
    image: icon,
    badge: icon,
    actions: [
      {
        action: "archive",
        title: "SPOTLIGHTS",
        url: "/",
        icon: "/static/icon.png",
      },
      {
        action: "hoge",
        title: "SPOTLIGHTSs",
        icon: "/static/icon.png",
      },
    ],
  };

  event.waitUntil(self.registration.showNotification(title, options));
});

self.addEventListener("notificationclick", function (event) {

  console.log(event.action);
  if (event.action === "archive") {
    // ユーザーが [Archive] アクションを選択しました。
    clients.openWindow("/");
    return
  }

  clients.openWindow("/");
  event.notification.close();
});
