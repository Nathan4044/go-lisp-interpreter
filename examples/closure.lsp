(def addBuilder (lambda (n) (lambda (m) (+ n m))))

(def addTwo (addBuilder 2))
(def addFive (addBuilder 5))

(def num 10)

(print (addTwo num))
(print (addFive num))
