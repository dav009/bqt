WITH TABLEA AS (
  SELECT column1 AS c1 FROM `dataset`.`table1`
)

SELECT column2, v FROM `dataset`.`table2` as DUMMY
LEFT JOIN TABLEA ON DUMMY.column2 =  TABLEA.c1
