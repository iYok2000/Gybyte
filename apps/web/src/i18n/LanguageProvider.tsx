"use client";

// Lightweight client-side i18n for the AdReady pages (TH/EN).
//
// - LanguageProvider holds the current language ("th" | "en"), defaults to "th",
//   and persists the choice to localStorage.
// - useLang() returns { lang, setLang, t, tList }.
// - When used OUTSIDE a provider, the context falls back to a default value with
//   lang "th" and a working `t`, so isolated component tests (which render
//   without a provider) get the byte-identical Thai strings.

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useSyncExternalStore,
  type ReactNode,
} from "react";

export type Lang = "th" | "en";

type DictValue = string | string[];
type Dictionary = Record<string, DictValue>;

const STORAGE_KEY = "adready-lang";

// -- Thai dictionary -------------------------------------------------------
// Values used by existing tests MUST stay byte-identical to the original
// hardcoded Thai strings.
const th: Dictionary = {
  // Nav
  "nav.toggle": "EN",
  "nav.toggleAria": "เปลี่ยนเป็นภาษาอังกฤษ",

  // Audit page
  "audit.page.title": "ตรวจสอบความพร้อม AdSense",
  "audit.page.subtitle":
    "กรอก URL ของเว็บไซต์เพื่อตรวจสอบความพร้อมตามเกณฑ์พื้นฐานของ Google AdSense",
  "audit.page.passed": "ผ่าน {passed}/{total}",

  // Audit form
  "audit.form.inputLabel": "URL ของเว็บไซต์ที่ต้องการตรวจสอบ",
  "audit.form.submit": "ตรวจสอบ",
  "audit.form.submitting": "กำลังตรวจสอบ...",
  "audit.form.error.empty": "กรุณากรอก URL ของเว็บไซต์ที่ต้องการตรวจสอบ",
  "audit.form.error.too-long": "URL ยาวเกินกำหนด (สูงสุด 2048 อักขระ)",
  "audit.form.error.invalid-format":
    "รูปแบบ URL ไม่ถูกต้อง กรุณากรอก URL ที่ถูกต้อง",
  "audit.form.error.bad-scheme":
    "รองรับเฉพาะ URL ที่ขึ้นต้นด้วย http:// หรือ https:// เท่านั้น",

  // Score display
  "audit.score.loading": "กำลังตรวจสอบ...",
  "audit.score.indeterminate": "ไม่สามารถประเมินคะแนนได้",
  "audit.score.invalid": "คะแนนไม่ถูกต้อง",
  "audit.score.caption": "คะแนนความพร้อม (เต็ม 100)",
  "audit.band.low": "ต้องปรับปรุง",
  "audit.band.mid": "เกือบพร้อม",
  "audit.band.high": "พร้อมแล้ว",

  // Results
  "audit.result.pass": "ผ่าน",
  "audit.result.fail": "ไม่ผ่าน",
  "audit.result.getHelp": "แก้ปัญหานี้",
  "audit.result.empty": "ไม่มีผลลัพธ์",

  // Error state
  "audit.error.rateLimit": "ส่งคำขอบ่อยเกินไป",
  "audit.error.generic": "การตรวจสอบไม่สำเร็จ",
  "audit.error.retry": "ลองใหม่อีกครั้ง",
  "audit.error.wait": "กรุณารออีก {n} วินาที",

  // Fix AdSense page
  "fix.title": "บริการแก้ปัญหาความพร้อม AdSense",
  "fix.intro":
    "ไม่ผ่านเกณฑ์บางข้อ? เราช่วยคุณแก้ไขให้เว็บไซต์พร้อมสำหรับ Google AdSense ไม่ว่าจะเลือกลงมือทำเองด้วยคู่มือของเรา หรือให้ทีมงานดูแลให้ครบทุกขั้นตอน",
  "fix.help.title": "เราช่วยอะไรได้บ้าง",
  "fix.help.body":
    "จากผลการตรวจสอบ เราจะช่วยแก้ไขทุกเกณฑ์ที่ไม่ผ่าน ตั้งแต่การติดตั้ง ads.txt, การเปิดใช้งาน HTTPS, การจัดการ robots.txt ไปจนถึงการเพิ่มหน้า Privacy Policy, Contact และ About Us ให้ครบถ้วนตามที่ AdSense คาดหวัง",
  "fix.recommended": "แนะนำ",
  "fix.back": "← กลับไปหน้าตรวจสอบ",

  "fix.pkg.diy.desc": "คู่มือลงมือทำเอง พร้อมขั้นตอนแก้ไขทีละข้อ",
  "fix.pkg.diy.cta": "รับคู่มือ",
  "fix.pkg.diy.features": [
    "คู่มือแก้ไขทุกเกณฑ์ AdSense",
    "เทมเพลต ads.txt และ robots.txt",
    "เช็กลิสต์หน้า Privacy / Contact / About",
  ],

  "fix.pkg.dfy.desc": "บริการติดตั้งและแก้ไขให้ครบทุกเกณฑ์โดยทีมงาน",
  "fix.pkg.dfy.cta": "เริ่มใช้บริการ",
  "fix.pkg.dfy.features": [
    "ทีมงานแก้ไขให้ทั้งหมด",
    "ตั้งค่า HTTPS, ads.txt, robots.txt",
    "เพิ่มหน้า compliance ที่จำเป็น",
    "ตรวจสอบซ้ำจนพร้อมยื่น AdSense",
  ],

  // Landing page — SEO headings (crawlable H1/H2, keyword-rich)
  "landing.seo.h1":
    "AdReady — เครื่องมือตรวจสอบความพร้อม Google AdSense ฟรี เช็กว่าเว็บผ่าน AdSense ไหม",
  "landing.seo.about.h2":
    "ตรวจสอบความพร้อม AdSense: ทำเว็บให้พร้อมและสมัคร AdSense ให้ผ่าน",
  "landing.seo.features.h2":
    "เครื่องมือตรวจ AdSense ฟรี: ตรวจ ads.txt, HTTPS และ robots.txt",

  // Landing page — Hero
  "landing.hero.description":
    "AdReady เครื่องมือตรวจสอบความพร้อม AdSense ฟรี เช็กว่าเว็บผ่าน AdSense ไหม — ตรวจ ads.txt, HTTPS, robots.txt และหน้า Privacy, Contact, About ที่ผู้ตรวจมองหา คลิกเดียวรู้คะแนน 0–100 พร้อมวิธีทำเว็บให้พร้อม AdSense",
  "landing.hero.cta": "เริ่มตรวจเว็บไซต์",

  // Landing page — About
  "landing.about.label": "ความพร้อมสำหรับ AdSense",
  "landing.about.heading": [
    "เราทำให้เว็บของคุณ ",
    "พร้อมสมัคร AdSense ให้ผ่าน",
    " ไม่ต้องเดา รู้คะแนนชัดเจน และสิ่งที่ต้องแก้แบบเจาะจง",
  ],
  "landing.about.body":
    "AdReady ตรวจสอบความพร้อม AdSense ด้วยเกณฑ์เดียวกับที่ผู้ตรวจ AdSense ให้ความสำคัญ — ตรวจ ads.txt, HTTPS, robots.txt และลิงก์หน้า Privacy Policy, Contact และ About บนหน้าแรก — แล้วสรุปเป็นคะแนนความพร้อมพร้อมวิธีแก้ที่นำไปทำเว็บให้พร้อม AdSense ได้ทันที",

  // Landing page — Features
  "landing.features.header1":
    "ครบทุกอย่างที่ต้องใช้เพื่อสมัคร AdSense ให้ผ่าน",
  "landing.features.header2": "ตรวจ ให้คะแนน แก้ไข แล้วยื่น AdSense ได้เลย",
  "landing.features.card1.caption": "เห็นความพร้อมของเว็บได้ในพริบตา",
  "landing.features.card2.title": "ตรวจทันที",
  "landing.features.card2.items": [
    "ตรวจ ads.txt ที่รากของเว็บ",
    "ตรวจ HTTPS และใบรับรอง TLS ที่ถูกต้อง",
    "ยืนยันว่ามี robots.txt ที่ถูกต้อง",
    "สแกนหน้าแรกหาลิงก์ที่จำเป็น",
  ],
  "landing.features.card3.title": "คะแนนความพร้อม",
  "landing.features.card3.items": [
    "คะแนนความพร้อม 0–100 ในตัวเดียว",
    "แยกสีผ่าน/ไม่ผ่านรายข้อ",
    "จัดลำดับสิ่งที่ควรแก้ก่อน",
  ],
  "landing.features.card4.title": "แนวทางแก้ไข",
  "landing.features.card4.items": [
    "คำแนะนำแก้ไขทีละขั้นชัดเจน",
    "เทมเพลต ads.txt และ robots.txt พร้อมใช้",
    "บริการติดตั้งให้ (ถ้าต้องการ)",
  ],
  "landing.features.learnMore": "ดูเพิ่มเติม",
};

// -- English dictionary ----------------------------------------------------
const en: Dictionary = {
  // Nav
  "nav.toggle": "TH",
  "nav.toggleAria": "Switch to Thai",

  // Audit page
  "audit.page.title": "Check your AdSense readiness",
  "audit.page.subtitle":
    "Enter a website URL to check its readiness against Google AdSense's baseline criteria.",
  "audit.page.passed": "{passed}/{total} passed",

  // Audit form
  "audit.form.inputLabel": "Website URL to audit",
  "audit.form.submit": "Check",
  "audit.form.submitting": "Checking...",
  "audit.form.error.empty": "Please enter the website URL you want to check.",
  "audit.form.error.too-long": "URL is too long (2048 characters max).",
  "audit.form.error.invalid-format":
    "Invalid URL format. Please enter a valid URL.",
  "audit.form.error.bad-scheme":
    "Only URLs starting with http:// or https:// are supported.",

  // Score display
  "audit.score.loading": "Checking...",
  "audit.score.indeterminate": "Unable to evaluate a score",
  "audit.score.invalid": "Invalid score",
  "audit.score.caption": "Readiness score (out of 100)",
  "audit.band.low": "Needs work",
  "audit.band.mid": "Almost ready",
  "audit.band.high": "Ready",

  // Results
  "audit.result.pass": "Passed",
  "audit.result.fail": "Failed",
  "audit.result.getHelp": "Fix this",
  "audit.result.empty": "No results",

  // Error state
  "audit.error.rateLimit": "Too many requests",
  "audit.error.generic": "Audit failed",
  "audit.error.retry": "Try again",
  "audit.error.wait": "Please wait {n} more seconds",

  // Fix AdSense page
  "fix.title": "AdSense readiness fix service",
  "fix.intro":
    "Failing a few criteria? We help get your site ready for Google AdSense — do it yourself with our guide, or let our team handle every step.",
  "fix.help.title": "How we help",
  "fix.help.body":
    "Based on your audit results, we fix every failing criterion — from installing ads.txt, enabling HTTPS and managing robots.txt, to adding the Privacy Policy, Contact and About Us pages AdSense expects.",
  "fix.recommended": "Recommended",
  "fix.back": "← Back to the audit",

  "fix.pkg.diy.desc": "A do-it-yourself guide with step-by-step fixes.",
  "fix.pkg.diy.cta": "Get the guide",
  "fix.pkg.diy.features": [
    "Guide covering every AdSense criterion",
    "ads.txt and robots.txt templates",
    "Privacy / Contact / About page checklist",
  ],

  "fix.pkg.dfy.desc": "We set everything up and fix every criterion for you.",
  "fix.pkg.dfy.cta": "Get started",
  "fix.pkg.dfy.features": [
    "Our team fixes everything",
    "Configure HTTPS, ads.txt, robots.txt",
    "Add the required compliance pages",
    "Re-audit until you're ready to apply",
  ],

  // Landing page — SEO headings (crawlable H1/H2, keyword-rich)
  "landing.seo.h1":
    "AdReady — free AdSense readiness checker to check your AdSense eligibility",
  "landing.seo.about.h2":
    "AdSense readiness checker: get approved for Google AdSense",
  "landing.seo.features.h2":
    "Free ads.txt checker and website AdSense audit tool",

  // Landing page — Hero
  "landing.hero.description":
    "AdReady is a free AdSense readiness checker — check your AdSense eligibility across ads.txt, HTTPS, robots.txt, and the Privacy, Contact and About pages reviewers look for. One click, a clear 0–100 score, and exactly what to fix to get approved for Google AdSense.",
  "landing.hero.cta": "Audit your site",

  // Landing page — About
  "landing.about.label": "AdSense readiness",
  "landing.about.heading": [
    "We get you ",
    "approved for Google AdSense",
    ". No guesswork — just a clear score and the exact fixes.",
  ],
  "landing.about.body":
    "AdReady is a website AdSense audit tool that runs the same baseline checks AdSense reviewers care about — ads.txt, HTTPS, robots.txt, and homepage links to Privacy Policy, Contact and About — then returns a single readiness score with fixes you can act on to get approved for Google AdSense.",

  // Landing page — Features
  "landing.features.header1":
    "Everything you need to get approved for Google AdSense.",
  "landing.features.header2": "Check. Score. Fix. Ship.",
  "landing.features.card1.caption": "See your readiness at a glance.",
  "landing.features.card2.title": "Instant Audit.",
  "landing.features.card2.items": [
    "Checks ads.txt at your site root",
    "Verifies HTTPS and a valid TLS certificate",
    "Confirms robots.txt is present and valid",
    "Scans your homepage for the required links",
  ],
  "landing.features.card3.title": "Readiness Score.",
  "landing.features.card3.items": [
    "A single 0–100 readiness score",
    "Color-coded pass/fail per check",
    "A prioritized list of what to fix first",
  ],
  "landing.features.card4.title": "Fix Guidance.",
  "landing.features.card4.items": [
    "Clear, step-by-step fix instructions",
    "Ready-made ads.txt and robots.txt templates",
    "Optional done-for-you service",
  ],
  "landing.features.learnMore": "Learn more",
};

const dictionaries: Record<Lang, Dictionary> = { th, en };

export type TFunc = (
  key: string,
  params?: Record<string, string | number>,
) => string;
export type TListFunc = (key: string) => string[];

function interpolate(
  template: string,
  params?: Record<string, string | number>,
): string {
  if (!params) return template;
  let out = template;
  for (const [name, value] of Object.entries(params)) {
    out = out.split(`{${name}}`).join(String(value));
  }
  return out;
}

function translate(
  lang: Lang,
  key: string,
  params?: Record<string, string | number>,
): string {
  const value = dictionaries[lang]?.[key] ?? th[key];
  if (typeof value === "string") return interpolate(value, params);
  // Missing key or an array requested via t(): fall back to the key itself.
  return key;
}

function translateList(lang: Lang, key: string): string[] {
  const value = dictionaries[lang]?.[key] ?? th[key];
  return Array.isArray(value) ? value : [];
}

export interface LanguageContextValue {
  lang: Lang;
  setLang: (lang: Lang) => void;
  t: TFunc;
  tList: TListFunc;
}

// Default context value — Thai, no-op setter, working translators. Ensures
// components rendered without a provider (unit tests) get Thai strings.
const defaultValue: LanguageContextValue = {
  lang: "th",
  setLang: () => {},
  t: (key, params) => translate("th", key, params),
  tList: (key) => translateList("th", key),
};

const LanguageContext = createContext<LanguageContextValue>(defaultValue);

// External store around localStorage. useSyncExternalStore consumes this so the
// language survives reloads without calling setState in an effect (avoids
// cascading renders) and without hydration mismatches: the server snapshot is
// always "th", and the client re-reads localStorage after hydration.
const langStore = {
  lang: "th" as Lang,
  hydrated: false,
  listeners: new Set<() => void>(),

  subscribe(listener: () => void): () => void {
    langStore.listeners.add(listener);
    return () => {
      langStore.listeners.delete(listener);
    };
  },

  getSnapshot(): Lang {
    // Lazily read the persisted choice on the first client-side read.
    if (!langStore.hydrated) {
      langStore.hydrated = true;
      try {
        const stored = localStorage.getItem(STORAGE_KEY);
        if (stored === "th" || stored === "en") {
          langStore.lang = stored;
        }
      } catch {
        // localStorage unavailable (SSR/private mode) — keep the default.
      }
    }
    return langStore.lang;
  },

  getServerSnapshot(): Lang {
    return "th";
  },

  set(next: Lang): void {
    langStore.lang = next;
    langStore.hydrated = true;
    try {
      localStorage.setItem(STORAGE_KEY, next);
    } catch {
      // Ignore persistence failures.
    }
    langStore.listeners.forEach((listener) => listener());
  },
};

export function LanguageProvider({ children }: { children: ReactNode }) {
  const lang = useSyncExternalStore(
    langStore.subscribe,
    langStore.getSnapshot,
    langStore.getServerSnapshot,
  );

  const setLang = useCallback((next: Lang) => {
    langStore.set(next);
  }, []);

  const value = useMemo<LanguageContextValue>(
    () => ({
      lang,
      setLang,
      t: (key, params) => translate(lang, key, params),
      tList: (key) => translateList(lang, key),
    }),
    [lang, setLang],
  );

  return (
    <LanguageContext.Provider value={value}>
      {children}
    </LanguageContext.Provider>
  );
}

export function useLang(): LanguageContextValue {
  return useContext(LanguageContext);
}

export default LanguageProvider;
