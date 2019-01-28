import datetime
from datetime import timedelta
from bs4 import BeautifulSoup
import requests
import pickle

page = 1
morePages = True
dates = []
while morePages:
	response = requests.get("http://www.thebearden.se/news?page={}".format(page))
	parsed_html = BeautifulSoup(response.text, "lxml")
	for node in parsed_html.body.find_all('div', attrs = {'id':'newsbox'}):
		str = node.contents[0].contents[0];
		node = node.find('p', attrs = {'class':'inline'})
		ddmmyy=node.contents[0].split('/')
		dates += [(datetime.date(int(ddmmyy[2]),int(ddmmyy[1]),int(ddmmyy[0])), str)]
	if not parsed_html.body.find('input', attrs = {'value':'Older'}):
		morePages = False
	page= page +1

now = datetime.date.today()
then = dates[-1][0]
then -= timedelta(days=then.weekday())
weeks = {}
weeks['{1}, Week {0:2d}'.format(then.isocalendar()[1] , then.isocalendar()[0])] = [0, ' | ']
while then<=now:
	weeks['{1}, Week {0:2d}'.format(then.isocalendar()[1] , then.isocalendar()[0])] = [0, ' | ']
	then+=timedelta(days=7)

for date in dates:
	weeks['{1}, Week {0:2d}'.format(date[0].isocalendar()[1] , date[0].isocalendar()[0])][0]+=1
	weeks['{1}, Week {0:2d}'.format(date[0].isocalendar()[1] , date[0].isocalendar()[0])][1]+=(date[1] + ' | ');

numweeks = {x: weeks[x][0] for x in weeks}
with open('weeks.pickle', 'wb') as f:
    pickle.dump(numweeks, f)

brokenstreak = False
currentstreak=0
beststreak=0
countingstreak=0
statistics=""

texttoprint = ""

for key, value in sorted(weeks.items(), reverse=True)[1:]:
	if(value[0] == 0):
		brokenstreak=True
		countingstreak=0
	else:
		countingstreak+=1
		if countingstreak > beststreak:
			beststreak=countingstreak
	if not brokenstreak:
		currentstreak +=1
	statistics+="{}: {}\n{}\n".format(key, value[0], value[1])

if(weeks['{1}, Week {0:2d}'.format(now.isocalendar()[1] , now.isocalendar()[0])][0]>0):
	currentstreak+=1
	if(currentstreak > beststreak):
		beststreak=currentstreak
else:
	texttoprint = "Current week pending.<br>"

print(texttoprint + "Current streak: {} weeks, best streak: {} weeks".format(currentstreak,beststreak))
print(statistics)
