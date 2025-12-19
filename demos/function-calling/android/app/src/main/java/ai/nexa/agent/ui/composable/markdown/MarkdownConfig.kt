package com.nexa.studio.ui.chat.markdown

import ai.nexa.agent.ui.theme.chatMessageAssistantText
import ai.nexa.agent.ui.theme.chatMessageUserBg
import ai.nexa.agent.util.chatMessageStyle
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.mikepenz.markdown.compose.components.MarkdownComponent
import com.mikepenz.markdown.compose.components.MarkdownComponentModel
import com.mikepenz.markdown.compose.components.markdownComponents
import com.mikepenz.markdown.compose.elements.MarkdownTable
import com.mikepenz.markdown.compose.elements.MarkdownTableHeader
import com.mikepenz.markdown.compose.elements.MarkdownTableRow
import com.mikepenz.markdown.m3.Markdown
import com.mikepenz.markdown.m3.markdownColor
import com.mikepenz.markdown.m3.markdownTypography
import com.mikepenz.markdown.model.MarkdownState
import com.mikepenz.markdown.model.markdownPadding

/**
 * Custom table component configuration that supports multiline display without ellipsis.
 */
fun customTableComponent(): MarkdownComponent = { model: MarkdownComponentModel ->
    MarkdownTable(
        content = model.content,
        node = model.node,
        style = model.typography.table,
        headerBlock = { content, header, tableWidth, style ->
            MarkdownTableHeader(
                content = content,
                header = header,
                tableWidth = tableWidth,
                style = style,
                maxLines = Int.MAX_VALUE,
                overflow = TextOverflow.Clip
            )
        },
        rowBlock = { content, row, tableWidth, style ->
            MarkdownTableRow(
                content = content,
                header = row,
                tableWidth = tableWidth,
                style = style,
                maxLines = Int.MAX_VALUE,
                overflow = TextOverflow.Clip
            )
        }
    )
}

/**
 * Custom Markdown component with unified configuration including custom table styles.
 *
 * @param markdownState The markdown state to render
 * @param modifier The modifier to apply to the component
 * @param textColor Text color, defaults to theme color if null
 * @param codeBackground Code background color, defaults to theme color if null
 * @param textStyle Text style, defaults to chatMessageStyle
 */
@Composable
fun CustomMarkdown(
    markdownState: MarkdownState,
    modifier: Modifier = Modifier,
    textColor: Color? = null,
    codeBackground: Color? = null,
    textStyle: TextStyle = chatMessageStyle,
) {
    val finalTextColor = textColor ?: MaterialTheme.colorScheme.chatMessageAssistantText
    val finalCodeBackground =
        codeBackground ?: MaterialTheme.colorScheme.chatMessageUserBg.copy(alpha = 0.3f)

    Markdown(
        markdownState = markdownState,
        colors = markdownColor(
            text = finalTextColor,
            codeBackground = finalCodeBackground
        ),
        typography = markdownTypography(
            text = textStyle,
            code = MaterialTheme.typography.bodySmall,
            h1 = MaterialTheme.typography.titleMedium,
            h2 = MaterialTheme.typography.titleSmall,
            h3 = MaterialTheme.typography.titleSmall
        ),
        padding = markdownPadding(block = 4.dp),
        components = markdownComponents(
            table = customTableComponent()
        ),
        modifier = modifier
    )
}

