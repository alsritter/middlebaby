import { createWebHistory, createRouter } from "vue-router";
import { RouteRecordRaw } from "vue-router";

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    alias: "/home",
    component: () => import("./views/MainContainer.vue"),
    children: [
      {
        path: "/case",
        component: () => import("./views/casemanager/CaseManager.vue"),
      },
      {
        path: "/capturemanager",
        component: () => import("./views/capturemanager/DataList.vue")
      },
      {
        path: "/setting",
        component: () => import("./views/setting/MBSetting.vue")
      }
    ]
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
