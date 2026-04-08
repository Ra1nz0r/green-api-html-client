// ===============================
// ПЕРЕКЛЮЧЕНИЕ ТЕМЫ
// ===============================

(function initTheme() {
  try {
    var savedTheme = localStorage.getItem("green_api_theme");
    if (savedTheme === "dark" || savedTheme === "light") {
      document.documentElement.setAttribute("data-theme", savedTheme);
      updateThemeButton(savedTheme);
    }
  } catch (e) {}
})();

function updateThemeButton(theme) {
  var btn = document.getElementById("themeToggle");
  if (!btn) return;
  btn.textContent = theme === "dark" ? "☀️" : "🌙";
}

document.getElementById("themeToggle").addEventListener("click", function () {
  var currentTheme = document.documentElement.getAttribute("data-theme") || "light";
  var nextTheme = currentTheme === "light" ? "dark" : "light";

  document.documentElement.setAttribute("data-theme", nextTheme);
  updateThemeButton(nextTheme);

  try {
    localStorage.setItem("green_api_theme", nextTheme);
  } catch (e) {}
});

// ===============================
// УТИЛИТЫ
// ===============================

var responseField = document.getElementById("response");
var responseStatus = document.getElementById("responseStatus");

function setStatus(text, isError) {
  responseStatus.textContent = text;
  responseStatus.classList.remove("status-success", "status-error");

  if (isError) {
    responseStatus.classList.add("status-error");
  } else {
    responseStatus.classList.add("status-success");
  }
}

function setResponse(data) {
  if (typeof data === "string") {
    responseField.value = data;
    return;
  }

  responseField.value = JSON.stringify(data, null, 2);
}

function safeParseJson(text) {
  try {
    return JSON.parse(text);
  } catch (e) {
    return text;
  }
}

function getConnectionParams() {
  return {
    idInstance: document.getElementById("idInstance").value.trim(),
    apiTokenInstance: document.getElementById("apiTokenInstance").value.trim()
  };
}

function buildQuery(params) {
  var searchParams = new URLSearchParams();
  Object.keys(params).forEach(function (key) {
    if (params[key] !== "") {
      searchParams.append(key, params[key]);
    }
  });
  return searchParams.toString();
}

async function requestJson(url, options) {
  setStatus("Выполняется запрос...", false);

  try {
    var response = await fetch(url, options || {});
    var text = await response.text();
    var data = safeParseJson(text);

    setResponse(data);

    if (!response.ok) {
      setStatus("Запрос завершился с ошибкой", true);
      return;
    }

    setStatus("Запрос выполнен успешно", false);
  } catch (error) {
    setResponse({
      error: "Network or unexpected error",
      details: String(error)
    });
    setStatus("Ошибка сети или выполнения запроса", true);
  }
}

// ===============================
// КНОПКИ GET-МЕТОДОВ
// ===============================

document.getElementById("getSettingsBtn").addEventListener("click", function () {
  var connection = getConnectionParams();

  var query = buildQuery({
    idInstance: connection.idInstance,
    apiTokenInstance: connection.apiTokenInstance
  });

  requestJson("/api/settings?" + query, {
    method: "GET"
  });
});

document.getElementById("getStateInstanceBtn").addEventListener("click", function () {
  var connection = getConnectionParams();

  var query = buildQuery({
    idInstance: connection.idInstance,
    apiTokenInstance: connection.apiTokenInstance
  });

  requestJson("/api/state?" + query, {
    method: "GET"
  });
});

// ===============================
// КНОПКА sendMessage
// ===============================

document.getElementById("sendMessageBtn").addEventListener("click", function () {
  var connection = getConnectionParams();

  var payload = {
    idInstance: connection.idInstance,
    apiTokenInstance: connection.apiTokenInstance,
    chatId: document.getElementById("messageChatId").value.trim(),
    message: document.getElementById("messageText").value.trim()
  };

  requestJson("/api/message", {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(payload)
  });
});

// ===============================
// КНОПКА sendFileByUrl
// ===============================

document.getElementById("sendFileBtn").addEventListener("click", function () {
  var connection = getConnectionParams();

  var payload = {
    idInstance: connection.idInstance,
    apiTokenInstance: connection.apiTokenInstance,
    chatId: document.getElementById("fileChatId").value.trim(),
    urlFile: document.getElementById("urlFile").value.trim(),
    fileName: document.getElementById("fileName").value.trim(),
    caption: document.getElementById("caption").value.trim()
  };

  requestJson("/api/file", {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(payload)
  });
});

// ===============================
// КНОПКА ОЧИСТКИ И КОПИРОВАНИЯ
// ===============================

document.getElementById("clearResponseBtn").addEventListener("click", function () {
  responseField.value = "";
  responseStatus.textContent = "Поле ответа очищено";
  responseStatus.classList.remove("status-success", "status-error");
});

document.getElementById("copyResponseBtn").addEventListener("click", async function () {
  try {
    await navigator.clipboard.writeText(responseField.value);
    setStatus("Ответ скопирован в буфер обмена", false);
  } catch (e) {
    setStatus("Не удалось скопировать ответ", true);
  }
});

// ===============================
// ПОКАЗ / СКРЫТИЕ ТОКЕНА
// ===============================

document.getElementById("showToken").addEventListener("change", function () {
  var tokenInput = document.getElementById("apiTokenInstance");
  tokenInput.type = this.checked ? "text" : "password";
});

// ===============================
// АККОРДЕОНЫ
// ===============================

document.querySelectorAll(".accordion__toggle").forEach(function (button) {
  button.addEventListener("click", function () {
    var accordion = this.closest(".accordion");
    accordion.classList.toggle("open");
  });
});