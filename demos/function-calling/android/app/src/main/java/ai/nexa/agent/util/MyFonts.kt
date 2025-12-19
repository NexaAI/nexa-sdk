package ai.nexa.agent.util

import ai.nexa.agent.R
import androidx.compose.foundation.layout.Column
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.Font
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.TextUnit
import androidx.compose.ui.unit.TextUnitType
import androidx.compose.ui.unit.sp

// 定义自定义字体家族
val fkGroteskNeueTrial = FontFamily(
    Font(R.font.fk_grotesk_neue_trial_black, FontWeight.Black, FontStyle.Normal),
    Font(R.font.fk_grotesk_neue_trial_black_italic, FontWeight.Black, FontStyle.Italic),
    // 加载字体文件，设置字体样式为正常，字体粗细为粗体
    Font(R.font.fk_grotesk_neue_trial_bold, FontWeight.Bold, FontStyle.Normal),
    Font(R.font.fk_grotesk_neue_trial_bold_italic, FontWeight.Bold, FontStyle.Italic),
    // 加载字体文件，设置字体样式为斜体，字体粗细为正常
    Font(R.font.fk_grotesk_neue_trial_italic, FontWeight.Normal, FontStyle.Italic),
    Font(R.font.fk_grotesk_neue_trial_light, FontWeight.Light, FontStyle.Normal),
    Font(R.font.fk_grotesk_neue_trial_light_italic, FontWeight.Light, FontStyle.Italic),
    Font(R.font.fk_grotesk_neue_trial_medium, FontWeight.Medium, FontStyle.Normal),
    Font(R.font.fk_grotesk_neue_trial_medium_italic, FontWeight.Medium, FontStyle.Italic),
    // 加载字体文件，设置字体样式为正常，字体粗细为正常
    Font(R.font.fk_grotesk_neue_trial_regular, FontWeight.Normal, FontStyle.Normal),
    Font(R.font.fk_grotesk_neue_trial_thin, FontWeight.Thin, FontStyle.Normal),
    Font(R.font.fk_grotesk_neue_trial_thin_italic, FontWeight.Thin, FontStyle.Italic),
)

val chatMessageStyle = TextStyle(
    fontFamily = fkGroteskNeueTrial,
    fontWeight = FontWeight(400),
    lineHeight = TextUnit(24f, TextUnitType.Sp),
    letterSpacing = TextUnit(0.25f, TextUnitType.Sp),
    fontSize = 16.sp,
)

val thinkMessageStyle = TextStyle(
    fontFamily = fkGroteskNeueTrial,
    fontWeight = FontWeight(400),
    lineHeight = TextUnit(22f, TextUnitType.Sp),
    letterSpacing = TextUnit(0.25f, TextUnitType.Sp),
    fontSize = 16.sp,
)

@Preview
@Composable
fun CustomFontExample() {
    // 使用 Text 组件显示文本，指定使用自定义字体家族
    val textStyle = TextStyle(

    )
    Column {
        Text(text = "This is a text using custom font.", fontStyle = FontStyle.Italic)
        Text(
            text = "This is a text using custom font.",
            fontFamily = fkGroteskNeueTrial,
            fontStyle = FontStyle.Italic
        )
    }

}