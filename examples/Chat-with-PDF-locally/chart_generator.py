import os
import matplotlib.pyplot as plt
import numpy as np

class ChartGenerator:
    def __init__(self, output_dir="charts"):
        self.output_dir = output_dir
        os.makedirs(output_dir, exist_ok=True)

    def _plot_line_or_column(self, chart_data_dict, chart_type):
        categories = chart_data_dict["chart_data"].get("categories", [])
        series = chart_data_dict["chart_data"].get("series", [])

        if not categories or not series:
            plt.text(0.5, 0.5, "No data to plot", ha="center", va="center", fontsize=16)
            return

        max_length = max(len(categories), max(len(s.get("values", [])) for s in series))
        x = np.arange(max_length)

        for s in series:
            values = s.get("values", [])
            values += [np.nan] * (max_length - len(values))
            if chart_type == "line":
                plt.plot(x, values, marker="o", label=s.get("name", ""))
            else:
                width = 0.8 / len(series)
                plt.bar(
                    x + series.index(s) * width,
                    values,
                    width=width,
                    label=s.get("name", ""),
                )

        plt.xlabel("Categories", fontsize=14)
        plt.ylabel("Values", fontsize=14)
        plt.legend(fontsize=10, loc="upper left", bbox_to_anchor=(1, 1))

        if len(categories) > 0:
            plt.xticks(
                x + (0.8 / len(series)) * (len(series) - 1) / 2,
                categories + [""] * (max_length - len(categories)),
                rotation=45,
                ha="right",
            )

    def _plot_pie(self, chart_data_dict):
        categories = chart_data_dict["chart_data"].get("categories", [])
        values = chart_data_dict["chart_data"].get("values", [])

        if not categories or not values:
            plt.text(0.5, 0.5, "No data to plot", ha="center", va="center", fontsize=16)
            return

        if len(categories) != len(values):
            min_length = min(len(categories), len(values))
            categories = categories[:min_length]
            values = values[:min_length]

        plt.pie(
            values, labels=categories, autopct="%1.1f%%", textprops={"fontsize": 10}
        )

    def _plot_bubble_or_scatter(self, chart_data_dict, chart_type):
        series = chart_data_dict["chart_data"].get("series", [])

        if not series:
            plt.text(0.5, 0.5, "No data to plot", ha="center", va="center", fontsize=16)
            return

        for s in series:
            data_points = s.get("data_points", [])
            x = [point.get("x", 0) for point in data_points]
            y = [point.get("y", 0) for point in data_points]

            if chart_type == "bubble":
                sizes = [point.get("size", 1) * 20 for point in data_points]
                plt.scatter(x, y, s=sizes, alpha=0.5, label=s.get("name", ""))
            else:
                plt.scatter(x, y, label=s.get("name", ""))

        plt.xlabel("X", fontsize=14)
        plt.ylabel("Y", fontsize=14)
        plt.legend(fontsize=10, loc="upper left", bbox_to_anchor=(1, 1))

    def plot_chart(self, chart_data_dict, chart_type=None):
        if chart_type is None:
            chart_type = chart_data_dict.get("chart_type", "").lower()
        else:
            chart_type = chart_type.lower()

        title = chart_data_dict.get("title_text", "")
        # description = chart_data_dict.get("description_text", "")

        plt.figure(figsize=(12, 7))
        plt.rcParams.update({"font.size": 12})

        if chart_type in ["line", "column_clustered", "clustered_column"]:
            self._plot_line_or_column(chart_data_dict, chart_type)
        elif chart_type == "pie":
            self._plot_pie(chart_data_dict)
        elif chart_type in ["bubble", "xy_scatter"]:
            self._plot_bubble_or_scatter(chart_data_dict, chart_type)
        else:
            plt.text(
                0.5,
                0.5,
                f"Unsupported chart type: {chart_type}",
                ha="center",
                va="center",
                fontsize=16,
            )

        plt.title(title, fontsize=20, pad=20)
        # plt.figtext(0.5, 0.01, description, wrap=True, ha="center", fontsize=10)
        plt.tight_layout()

        filename = f"{chart_type}_{title.replace(' ', '_')}.png"
        filepath = os.path.join(self.output_dir, filename)
        plt.savefig(filepath, dpi=300, bbox_inches="tight")
        plt.close()

        return filepath

    def plot_multiple_charts(self, chart_data_array):
        filepaths = []
        for chart_data in chart_data_array:
            if chart_data["type"] == "chart":
                filepath = self.plot_chart(chart_data)
                filepaths.append(filepath)
            elif chart_data["type"] == "text":
                # For text slides, we're not generating a file, so we'll skip
                pass
        return filepaths
