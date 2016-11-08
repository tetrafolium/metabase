(ns ring.middleware.test.multipart-params
  (:require [clojure.test :refer :all]
            [ring.middleware.multipart-params :refer :all]
            [ring.util.io :refer [string-input-stream]])
  (:import [java.io InputStream]))

(defn string-store [item]
  (-> (select-keys item [:filename :content-type])
      (assoc :content (slurp (:stream item)))))

(deftest test-wrap-multipart-params
  (let [form-body (str "--XXXX\r\n"
                       "Content-Disposition: form-data;"
                       "name=\"upload\"; filename=\"test.txt\"\r\n"
                       "Content-Type: text/plain\r\n\r\n"
                       "foo\r\n"
                       "--XXXX\r\n"
                       "Content-Disposition: form-data;"
                       "name=\"baz\"\r\n\r\n"
                       "qux\r\n"
                       "--XXXX--")
        handler (wrap-multipart-params identity {:store string-store})
        request {:headers {"content-type" "multipart/form-data; boundary=XXXX"
                           "content-length" (str (count form-body))}
                 :params {"foo" "bar"}
                 :body (string-input-stream form-body)}
        response (handler request)]
    (is (= (get-in response [:params "foo"]) "bar"))
    (is (= (get-in response [:params "baz"]) "qux"))
    (let [upload (get-in response [:params "upload"])]
      (is (= (:filename upload)     "test.txt"))
      (is (= (:content-type upload) "text/plain"))
      (is (= (:content upload)      "foo")))))

(deftest test-multiple-params
  (let [form-body (str "--XXXX\r\n"
                       "Content-Disposition: form-data;"
                       "name=\"foo\"\r\n\r\n"
                       "bar\r\n"
                       "--XXXX\r\n"
                       "Content-Disposition: form-data;"
                       "name=\"foo\"\r\n\r\n"
                       "baz\r\n"
                       "--XXXX--")
        handler (wrap-multipart-params identity {:store string-store})
        request {:headers {"content-type" "multipart/form-data; boundary=XXXX"
                           "content-length" (str (count form-body))}
                 :body (string-input-stream form-body)}
        response (handler request)]
    (is (= (get-in response [:params "foo"])
           ["bar" "baz"]))))

(defn all-threads []
  (.keySet (Thread/getAllStackTraces)))

(deftest test-multipart-threads
  (testing "no thread leakage when handler called"
    (let [handler (wrap-multipart-params identity)]
      (dotimes [_ 200]
        (handler {}))
      (is (< (count (all-threads))
             100))))

  (testing "no thread leakage from default store"
    (let [form-body (str "--XXXX\r\n"
                         "Content-Disposition: form-data;"
                         "name=\"upload\"; filename=\"test.txt\"\r\n"
                         "Content-Type: text/plain\r\n\r\n"
                         "foo\r\n"
                         "--XXXX--")]
      (dotimes [_ 200]
        (let [handler (wrap-multipart-params identity)
              request {:headers {"content-type" "multipart/form-data; boundary=XXXX"
                                 "content-length" (str (count form-body))}
                       :body (string-input-stream form-body)}]
          (handler request))))
    (is (< (count (all-threads))
           100))))

(deftest wrap-multipart-params-cps-test
  (let [handler   (wrap-multipart-params (fn [req respond _] (respond req)))
        form-body (str "--XXXX\r\n"
                       "Content-Disposition: form-data;"
                       "name=\"foo\"\r\n\r\n"
                       "bar\r\n"
                       "--XXXX--")
        request   {:headers {"content-type" "multipart/form-data; boundary=XXXX"}
                   :body    (string-input-stream form-body "UTF-8")}
        response  (promise)
        exception (promise)]
        (handler request response exception)
        (is (= (get-in @response [:multipart-params "foo"]) "bar"))
        (is (not (realized? exception)))))

(deftest multipart-params-request-test
  (is (fn? multipart-params-request)))

(deftest test-utf8-encoding-support
  (let [form-body (str "--XXXX\r\n"
                       "Content-Disposition: form-data;"
                       "name=\"foo\"\r\n\r\n"
                       "Øæßç®£èé\r\n"
                       "--XXXX--")
        request {:headers {"content-type"
                           (str "multipart/form-data; boundary=XXXX")}
                 :body (string-input-stream form-body "UTF-8")}
        request* (multipart-params-request request)]
        (is (= (get-in request* [:multipart-params "foo"]) "Øæßç®£èé"))))
