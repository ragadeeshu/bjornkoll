import pickle
import json
from statistics import mean

weeks = {}
with open('weeks.pickle','rb') as f:
    weeks = pickle.load(f)

tobejson = {}
tobejson["weeks"] = sorted(weeks.items(), reverse=False)[:-1]
# print(mean([x[1] for x in tobejson["weeks"][-10:]]))
tobejson["tenweekaverage"] = mean([x[1] for x in tobejson["weeks"][-10:]])
print(json.dumps(tobejson))
