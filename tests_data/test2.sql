WITH TABLEA AS (
  SELECT column1 AS c1 FROM `dataset`.`table1`
)

SELECT column2, v FROM `dataset`.`table2`
LEFT JOIN TABLEA ON table2.column2 =  TABLEA.c1
