import time

# mock tools

def get_weather(location):
    # get weather
    return f"location: {location} weather: 18℃"

def send_email(content):
    # send email
    return f"email send: content: {content}"

def timenow(unuse):
    # get time
    return f"{time.strftime("%a %b %d %H:%M:%S %Y", time.localtime())}"

def finished():
    # finished
    return "done"
