(def reduce (lambda (lst fn acc)
              (if (= 0 (len lst))
                acc 
                (reduce (rest lst) fn (fn acc (first lst))))))

(def map (lambda (lst fn)
           (reduce lst (lambda (acc n) (push acc (fn n))) '())))

(def printList (lambda (l)
                 (print (str "List is " l))))

(def lst '(1 2 3 4 5))

(printList lst)
(printList (map lst (lambda (n) (* n n))))
