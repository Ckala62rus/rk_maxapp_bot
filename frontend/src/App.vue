<template>
  <div class="app">
    <!-- Header -->
    <header class="app-header">
      <div>
        <div class="brand">MAX • Поиск партий</div>
        <div class="muted">Mini App для поиска партий и контроля доступа</div>
      </div>
      <div class="status" v-if="user">
        <span class="pill" :class="user.isBlocked ? 'danger' : 'success'">
          {{ user.isBlocked ? "Доступ заблокирован" : "Пользователь" }}
        </span>
        <span class="pill" v-if="user.isAdmin">Админ</span>
      </div>
    </header>

<!--    <p>testInitData {{testInitData}}</p>-->

    <!-- InitData validation spinner -->
    <section class="card" v-if="loading">
      <div class="loading">
        <div class="spinner"></div>
        <div>
          <div class="loading-title">Проверяем доступ…</div>
          <div class="muted">Валидация initData MAX.</div>
        </div>
      </div>
    </section>

    <!-- Global error block -->
    <section class="card" v-if="error">
      <h2>Ошибка</h2>
      <p class="muted">{{ error }}</p>
      <p class="muted">Проверьте initData и настройки бэкенда.</p>
      <p class="muted">Debug: initData length = {{ initData.length }}</p>
      <div class="form">
        <label>
          InitData (dev)
          <textarea
            v-model="manualInitData"
            rows="3"
            placeholder="Вставьте строку initData из /api/dev/init-data"
          ></textarea>
        </label>
        <button class="btn primary" @click="useManualInitData">
          Использовать initData
        </button>
        <button class="btn ghost" @click="generateInitData">
          Сгенерировать initData (dev)
        </button>
      </div>
    </section>

    <!-- Profile form (first/last name) -->
    <section class="card" v-if="user && !user.profileComplete">
      <h2>Введите ФИО</h2>
      <p class="muted">
        Для доступа к поиску нужно заполнить имя и фамилию. После этого дождитесь
        подтверждения администратора.
      </p>
      <div class="form">
        <label>
          Имя
          <input v-model="profile.firstName" type="text" placeholder="Иван" />
        </label>
        <label>
          Фамилия
          <input v-model="profile.lastName" type="text" placeholder="Иванов" />
        </label>
        <button class="btn primary" :disabled="profileSaving" @click="saveProfile">
          {{ profileSaving ? "Сохраняем…" : "Отправить" }}
        </button>
      </div>
    </section>

    <!-- Waiting for admin approval -->
    <section class="card" v-if="user && user.profileComplete && !user.isApproved">
      <h2>Ожидает подтверждения</h2>
      <p class="muted">
        Администратор проверит данные и откроет доступ. Пока поиск недоступен.
      </p>
    </section>

    <!-- Blocked user notice -->
    <section class="card" v-if="user && user.isBlocked">
      <h2>Доступ заблокирован</h2>
      <p class="muted">Обратитесь к администратору для восстановления доступа.</p>
    </section>

    <!-- Search UI -->
    <section class="card" v-if="user && user.isApproved && !user.isBlocked">
      <h2>Поиск партий</h2>
      <div class="search">
        <input
          v-model="search.code"
          type="text"
          placeholder="Введите номер партии, например 21%12345 или 12345*"
        />
        <button class="btn primary" :disabled="search.loading" @click="searchBatches">
          {{ search.loading ? "Идет запрос…" : "Найти" }}
        </button>
      </div>

      <div class="inline-loading" v-if="search.loading">
        <div class="spinner small"></div>
        <span class="muted">Поиск партии…</span>
      </div>

      <div v-if="search.error" class="error-box">
        {{ search.error }}
      </div>

      <div v-if="search.results.length" class="results">
        <div class="result-card" v-for="item in search.results" :key="item.batch">
          <div class="result-title">{{ item.batch }}</div>
          <div class="muted">{{ item.namealias }}</div>
          <div class="result-grid">
            <div>
              <div class="label">Ячейки</div>
              <div>{{ item.wmslocation }}</div>
            </div>
            <div>
              <div class="label">Лицензии</div>
              <div>{{ item.license }}</div>
            </div>
            <div v-if="item.colorid">
              <div class="label">Цвет</div>
              <div>{{ item.colorid }}</div>
            </div>
            <div v-if="item.configid">
              <div class="label">Конфиг</div>
              <div>{{ item.configid }}</div>
            </div>
            <div v-if="item.username">
              <div class="label">ФИО</div>
              <div>{{ item.username }}</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Admin users management -->
    <section class="card admin" v-if="user && user.isAdmin && !user.isBlocked">
      <div class="admin-head">
        <div>
          <h2>Админ: пользователи</h2>
          <div class="muted">Поиск по ФИО пользователя</div>
        </div>
        <div class="admin-search">
          <input
            v-model="adminSearch"
            type="text"
            placeholder="Введите ФИО"
            @keyup.enter="loadUsers"
          />
          <button class="btn ghost" @click="loadUsers" :disabled="adminLoading">
            {{ adminLoading ? "Ищем…" : "Найти" }}
          </button>
        </div>
      </div>

      <div v-if="adminError" class="error-box">{{ adminError }}</div>

      <div class="admin-list" v-if="adminUsers.length">
        <div class="admin-row" v-for="u in adminUsers" :key="u.id">
          <div>
            <div class="result-title">
              {{ u.firstName || "Без имени" }} {{ u.lastName || "" }}
            </div>
            <div class="muted">MAX ID: {{ u.maxUserID }}</div>
          </div>
          <div class="admin-tags">
            <span class="pill" :class="u.isApproved ? 'success' : 'neutral'">
              {{ u.isApproved ? "Одобрен" : "Ожидает" }}
            </span>
            <span class="pill" :class="u.isBlocked ? 'danger' : 'neutral'">
              {{ u.isBlocked ? "Блок" : "Активен" }}
            </span>
            <span class="pill" v-if="u.isAdmin">Админ</span>
          </div>
          <div class="admin-actions">
            <button class="btn ghost" @click="updateUser(u.id, { isApproved: true })">
              Одобрить
            </button>
            <button class="btn ghost" @click="updateUser(u.id, { isBlocked: true })">
              Блок
            </button>
            <button class="btn ghost" @click="updateUser(u.id, { isBlocked: false })">
              Разблок
            </button>
            <button
              class="btn ghost"
              v-if="!u.isAdmin"
              @click="updateUser(u.id, { isAdmin: true })"
            >
              Сделать админом
            </button>
            <button
              class="btn ghost"
              v-else
              @click="updateUser(u.id, { isAdmin: false })"
            >
              Снять админа
            </button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
// Vue utilities for state and lifecycle.
import { onMounted, reactive, ref } from "vue";

// Global loading and error flags for initData validation.
const loading = ref(true);
const error = ref("");
// Authenticated user info from backend.
const user = ref(null);
// Raw initData string (used for auth in API calls).
const initData = ref("");
// Manual initData input for dev testing.
const manualInitData = ref("");

// const testInitData = ref(null);

// API base allows bypassing Vite proxy when needed.
const apiBase = (import.meta.env.VITE_API_BASE || "").replace(/\/$/, "");
const apiUrl = (path) => (apiBase ? `${apiBase}${path}` : path);

// Profile form state.
const profile = reactive({
  firstName: "",
  lastName: "",
});
const profileSaving = ref(false);

// Search state and results.
const search = reactive({
  code: "",
  loading: false,
  error: "",
  results: [],
});

// Admin list and actions state.
const adminUsers = ref([]);
const adminLoading = ref(false);
const adminError = ref("");
const adminSearch = ref("");

// getInitData reads MAX initData from URL, hash, localStorage, WebApp or env.
const getInitData = () => {
  // Allow quick testing via URL query param (initData / init_data).
  const url = new URL(window.location.href);
  const initDataFromUrl =
    url.searchParams.get("initData") ||
    url.searchParams.get("init_data") ||
    url.searchParams.get("initdata");
  if (initDataFromUrl) {
    // Store for future reloads without query string.
    localStorage.setItem("maxapp_init_data", initDataFromUrl);
    return initDataFromUrl;
  }

  // Support direct params: auth_date/hash/user/query_id (dev helper).
  if (url.searchParams.get("auth_date") && url.searchParams.get("hash")) {
    const params = new URLSearchParams();
    ["auth_date", "query_id", "user", "hash"].forEach((key) => {
      const value = url.searchParams.get(key);
      if (value) {
        params.set(key, value);
      }
    });
    const rebuilt = params.toString();
    if (rebuilt) {
      localStorage.setItem("maxapp_init_data", rebuilt);
      return rebuilt;
    }
  }

  // Support hash-based params (e.g. #/path?initData=...).
  if (window.location.hash.includes("initData")) {
    const hash = window.location.hash.replace(/^#/, "");
    const hashUrl = new URL(`https://local/${hash}`);
    const initDataFromHash =
      hashUrl.searchParams.get("initData") ||
      hashUrl.searchParams.get("init_data") ||
      hashUrl.searchParams.get("initdata");
    if (initDataFromHash) {
      localStorage.setItem("maxapp_init_data", initDataFromHash);
      return initDataFromHash;
    }
  }

  // Fallback to previously stored value.
  const stored = localStorage.getItem("maxapp_init_data");
  if (stored) {
    return stored;
  }

  if (window?.WebApp?.InitData) {
    return window.WebApp.InitData;
  }
  if (window?.WebApp?.initData) {
    return window.WebApp.initData;
  }
  if (window?.Telegram?.WebApp?.initData) {
    return window.Telegram.WebApp.initData;
  }
  return import.meta.env.VITE_INIT_DATA || "";
};

// readJson safely reads JSON, even if response is empty.
const readJson = async (res) => {
  const text = await res.text();
  if (!text) {
    return null;
  }
  try {
    return JSON.parse(text);
  } catch (err) {
    return { success: false, message: text };
  }
};

// validate calls backend to validate initData and load user flags.
const validate = async () => {
  loading.value = true;
  error.value = "";
  initData.value = getInitData();

  if (!initData.value) {
    // In dev mode try to auto-generate initData from backend helper.
    if (import.meta.env.DEV) {
      try {
        const res = await fetch(
          apiUrl(`/api/dev/init-data?front=${encodeURIComponent(window.location.origin)}`)
        );
        const payload = await readJson(res);
        if (res.ok && payload?.data?.initData) {
          initData.value = payload.data.initData;
          localStorage.setItem("maxapp_init_data", initData.value);
        }
      } catch (err) {
        // Ignore and show original error below.
      }
    }
  }

  if (!initData.value) {
    error.value = "Не получены initData из MAX.";
    loading.value = false;
    return;
  }

  try {
    // Validate initData on backend.
    const res = await fetch(apiUrl("/api/auth/validate"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ initData: initData.value }),
    });
    const payload = await readJson(res);
    if (!res.ok || !payload?.success) {
      throw new Error(payload?.message || `Ошибка валидации (${res.status})`);
    }
    user.value = payload.data;
    profile.firstName = payload.data.firstName || "";
    profile.lastName = payload.data.lastName || "";
    // Load admin users list if search query is already set.
    if (user.value.isAdmin && adminSearch.value.trim()) {
      loadUsers();
    }
  } catch (err) {
    error.value = err.message || "Ошибка валидации";
  } finally {
    loading.value = false;
  }
};

// useManualInitData stores initData and re-runs validation.
const useManualInitData = () => {
  const value = manualInitData.value.trim();
  if (!value) {
    // If initData is already present, just re-validate it.
    if (initData.value) {
      validate();
      return;
    }
    error.value = "InitData пустая";
    return;
  }
  localStorage.setItem("maxapp_init_data", value);
  validate();
};

// generateInitData calls backend dev endpoint and stores result.
const generateInitData = async () => {
  try {
    const res = await fetch(
      apiUrl(`/api/dev/init-data?front=${encodeURIComponent(window.location.origin)}`)
    );
    const payload = await readJson(res);
    if (!res.ok || !payload?.data?.initData) {
      throw new Error(payload?.message || "Не удалось сгенерировать initData");
    }
    manualInitData.value = payload.data.initData;
    localStorage.setItem("maxapp_init_data", payload.data.initData);
    validate();
  } catch (err) {
    error.value = err.message || "Ошибка генерации initData";
  }
};

// saveProfile sends first/last name to backend.
const saveProfile = async () => {
  profileSaving.value = true;
  error.value = "";
  try {
    // Authorized request with initData header.
    const res = await fetch(apiUrl("/api/users/profile"), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Max-InitData": initData.value,
      },
      body: JSON.stringify({
        firstName: profile.firstName,
        lastName: profile.lastName,
      }),
    });
    const payload = await readJson(res);
    if (!res.ok || !payload?.success) {
      throw new Error(payload?.message || `Не удалось сохранить (${res.status})`);
    }
    // Mark profile as completed on UI side.
    user.value.profileComplete = true;
  } catch (err) {
    error.value = err.message || "Ошибка сохранения";
  } finally {
    profileSaving.value = false;
  }
};

// searchBatches executes DAX lookup and shows results.
const searchBatches = async () => {
  search.loading = true;
  search.error = "";
  search.results = [];
  try {
    // Pass initData header for auth.
    const res = await fetch(
      apiUrl(`/api/warehouse/batches?code=${encodeURIComponent(search.code)}`),
      {
        headers: {
          "X-Max-InitData": initData.value,
        },
      }
    );
    const payload = await readJson(res);
    if (!res.ok || !payload?.success) {
      throw new Error(payload?.message || `Ничего не найдено (${res.status})`);
    }
    search.results = payload.data || [];
  } catch (err) {
    search.error = err.message || "Ошибка запроса";
  } finally {
    search.loading = false;
  }
};

// loadUsers fetches admin list of users.
const loadUsers = async () => {
  const query = adminSearch.value.trim();
  if (!query) {
    adminUsers.value = [];
    adminError.value = "Введите ФИО для поиска";
    return;
  }
  adminLoading.value = true;
  adminError.value = "";
  try {
    const res = await fetch(apiUrl(`/api/admin/users?q=${encodeURIComponent(query)}`), {
      headers: {
        "X-Max-InitData": initData.value,
      },
    });
    const payload = await readJson(res);
    if (!res.ok || !payload?.success) {
      throw new Error(payload?.message || `Не удалось получить пользователей (${res.status})`);
    }
    adminUsers.value = payload.data || [];
  } catch (err) {
    adminError.value = err.message || "Ошибка загрузки";
  } finally {
    adminLoading.value = false;
  }
};

// updateUser applies admin changes to flags (approve/block/admin).
const updateUser = async (id, flags) => {
  adminError.value = "";
  try {
    const res = await fetch(apiUrl(`/api/admin/users/${id}`), {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
        "X-Max-InitData": initData.value,
      },
      body: JSON.stringify(flags),
    });
    const payload = await readJson(res);
    if (!res.ok || !payload?.success) {
      throw new Error(payload?.message || `Не удалось обновить (${res.status})`);
    }
    await loadUsers();
  } catch (err) {
    adminError.value = err.message || "Ошибка обновления";
  }
};

// Run initData validation on first load.
onMounted(() => {
  validate();
  // testInitData.value = window.WebApp.initData
});
</script>
