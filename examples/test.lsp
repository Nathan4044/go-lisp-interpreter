(def reduce (lambda (lst fn acc)
              (if (= 0 (len lst))
                acc 
                (reduce (rest lst) fn (fn acc (first lst))))))

(def map (lambda (lst fn)
           (reduce lst (lambda (acc n) (push acc (fn n))) '())))

(def printList (lambda (l)
                 (print (str "List is " l))))

(def rangeBuilder (lambda (lst start end) (if (= start end) lst (rangeBuilder (push lst start) (+ 1 start) end))))
(def range (lambda (n) (rangeBuilder '() 0 n)))

(def lst (range 1000))

(printList lst)
(printList (map lst (lambda (n) (- (* n n) (/ n 2)))))
