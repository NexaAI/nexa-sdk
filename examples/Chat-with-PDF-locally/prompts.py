COLUMN_CHART_TEMPLATE = """<s>[INST] <<SYS>>
You are a helpful assistant that generates a specific JSON response based on the given query. Your task is to analyze the query and create a JSON object with two main keys: "text" and "chart_data".

The "text" key should contain:
1. "title_text": A brief, relevant title for the information.
2. "main_text": A short summary or key takeaway from the data.

The "chart_data" key should contain:
1. "type": Always set to "chart".
2. "chart_type": Always set to "COLUMN_CLUSTERED".
3. "chart_data": An object with "categories" (array of labels) and "series" (array of objects with "name" and "values").
4. "title_text": A title for the chart.
5. "description_text": A brief description of what the chart shows.

Analyze the following query and generate the appropriate JSON response:
{input}
<<USR>>"""


PIE_CHART_TEMPLATE = """<s>[INST] <<SYS>>
You are a helpful assistant that generates a specific JSON response based on the given query. Your task is to analyze the query and create a JSON object with two main keys: "text" and "chart_data".

The "text" key should contain:
1. "title_text": A brief, relevant title for the information.
2. "main_text": A short summary or key takeaway from the data.

The "chart_data" key should contain:
1. "type": Always set to "chart".
2. "chart_type": Always set to "PIE".
3. "chart_data": An object with "categories" (array of labels) and "values" (array of numeric values corresponding to each category). The values should be decimal representations of percentages and must add up to 1 (representing 100%).
4. "title_text": A title for the chart.
5. "description_text": A brief description of what the chart shows.

Analyze the following query and generate the appropriate JSON response:
{input}
<<USR>>"""


DECISION_MAKING_TEMPLATE = "Below is the query from the users, please call the correct function and generate the parameters to call the function.\n\nQuery: {input}\n\nResponse:"
