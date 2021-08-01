SELECT count(*) from transactions where lat IS NULL;

SELECT t1.department_code, perdep, errperdep, errperdep/perdep
FROM
(select department_code,count(*) as perdep
from transactions group by department_code) t1
LEFT JOIN
(select department_code,count(*) as errperdep 
from transactions where nb_room = 0 group by department_code) t2
ON (t1.department_code = t2.department_code);
