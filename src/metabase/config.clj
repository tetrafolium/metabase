(ns metabase.config
  (:require (clojure.java [io :as io]
                          [shell :as shell])
            [clojure.string :as s]
            [environ.core :as environ]
            [medley.core :as m])
  (:import clojure.lang.Keyword))

(def ^:private ^:const app-defaults
  "Global application defaults"
  {;; Database Configuration  (general options?  dburl?)
   :mb-run-mode "prod"
   :mb-db-type "h2"
   ;:mb-db-dbname "postgres"
   ;:mb-db-host "localhost"
   ;:mb-db-port "5432"
   ;:mb-db-user "metabase"
   ;:mb-db-pass "metabase"
   :mb-db-file "metabase.db"
   :mb-db-automigrate "true"
   :mb-db-logging "true"
   ;; Embedded Jetty Webserver
   ;; check here for all available options:
   ;; https://github.com/ring-clojure/ring/blob/master/ring-jetty-adapter/src/ring/adapter/jetty.clj
   :mb-jetty-port "3000"
   ;; Other Application Settings
   :mb-password-complexity "normal"
   ;:mb-password-length "8"
   :max-session-age "20160"})                    ; session length in minutes (14 days)


(defn config-str
  "Retrieve value for a single configuration key.  Accepts either a keyword or a string.

   We resolve properties from these places:
   1.  environment variables (ex: MB_DB_TYPE -> :mb-db-type)
   2.  jvm opitons (ex: -Dmb.db.type -> :mb-db-type)
   3.  hard coded `app-defaults`"
  [k]
  (let [k (keyword k)]
    (or (k environ/env) (k app-defaults))))


;; These are convenience functions for accessing config values that ensures a specific return type
(defn ^Integer config-int  [k] (some-> k config-str Integer/parseInt))
(defn ^Boolean config-bool [k] (some-> k config-str Boolean/parseBoolean))
(defn ^Keyword config-kw   [k] (some-> k config-str keyword))


(def ^:const config-all
  "Global application configuration as a dictionary.
   Combines hard coded defaults with optional user specified overrides from environment variables."
  (into {} (for [k (keys app-defaults)]
               [k (config-str k)])))


(defn config-match
  "Retrieves all configuration values whose key begin with a specified prefix.
   The returned map will strip the prefix from the key names.
   All returned values will be Strings.

   For example if you wanted all of Java's internal environment config values:
   * (config-match \"java-\") -> {:version \"25.25-b02\" :info \"mixed mode\"}"
  [prefix]
  (let [prefix-regex (re-pattern (str ":" prefix ".*"))]
    (->> (merge
          (m/filter-keys (fn [k] (re-matches prefix-regex (str k))) app-defaults)
          (m/filter-keys (fn [k] (re-matches prefix-regex (str k))) environ/env))
      (m/map-keys (fn [k] (let [kstr (str k)] (keyword (subs kstr (+ 1 (count prefix))))))))))

(defn ^Boolean is-dev?  [] (= :dev  (config-kw :mb-run-mode)))
(defn ^Boolean is-prod? [] (= :prod (config-kw :mb-run-mode)))
(defn ^Boolean is-test? [] (= :test (config-kw :mb-run-mode)))


;;; Version stuff
;; Metabase version is of the format `GIT-TAG (GIT-SHORT-HASH GIT-BRANCH)`

(defn- version-info-from-shell-script []
  {:long  (-> (shell/sh "./version")           :out s/trim)
   :short (-> (shell/sh "./version" "--short") :out s/trim)})

(defn- version-info-from-properties-file []
  (with-open [reader (io/reader (io/resource "version.properties"))]
    (let [props (java.util.Properties.)]
      (.load props reader)
      (into {} (for [[k v] props]
                 [(keyword k) v])))))

(defn mb-version-info
  "Return information about the current version of Metabase.
   This comes from `resources/version.properties` for prod builds and is fetched from `git` via the `./version` script for dev.

     (mb-version) -> {:long \"v0.11.1 (6509c49 master)\", :short \"v0.11.1\"}"
  []
  (if (is-prod?)
    (version-info-from-properties-file)
    (version-info-from-shell-script)))