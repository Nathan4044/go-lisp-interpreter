(def rangeBuilder (lambda (lst start end) (if (= start end) lst (rangeBuilder (push! lst start) (+ 1 start) end))))
(def range (lambda (n) (rangeBuilder '() 0 n)))

(def reduce (lambda (lst fn acc)
              (if (= 0 (len lst))
                acc 
                (reduce (rest lst) fn (fn acc (first lst))))))

(def map (lambda (lst fn)
           (reduce lst (lambda (acc n) (push! acc (fn n))) '())))

(def fibIter (lambda (a b n) (if (= n 0) b (fibIter b (+ a b) (- n 1)))))

(def fib (lambda (n) (fibIter 0 1 n)))

(def lst (range 92))

(def result (map lst fib))

(print result)
