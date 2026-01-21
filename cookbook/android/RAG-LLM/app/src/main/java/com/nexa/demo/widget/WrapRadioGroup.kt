package com.nexa.demo.widget

import android.content.Context
import android.util.AttributeSet
import android.view.View
import android.widget.RadioGroup
import androidx.core.view.size
import com.nexa.demo.R
import kotlin.math.max

class WrapRadioGroup : RadioGroup {
    private var mMaxWidth = 0
    private var childMaxHeight = 80
    private var firstLinePadding = 0

    constructor(context: Context?) : super(context)

    constructor(context: Context?, attrs: AttributeSet?) : super(context, attrs)

    override fun onMeasure(widthMeasureSpec: Int, heightMeasureSpec: Int) {
        val widthMode = MeasureSpec.getMode(widthMeasureSpec)
        val widthSize = MeasureSpec.getSize(widthMeasureSpec)
        val heightMode = MeasureSpec.getMode(heightMeasureSpec)
        val heightSize = MeasureSpec.getSize(heightMeasureSpec)

        mMaxWidth = widthSize - getPaddingLeft() - getPaddingRight()

        val childCount = size
        var linePosition = 0
        var curLineWidth = 0
        var firstLineChildWidth = 0
        var firstLineChildCount = 0
        for (i in 0..<childCount) {
            val child = getChildAt(i)
            if (child.visibility != View.VISIBLE) {
                continue
            }
            val lp = child.layoutParams as LayoutParams
            val childWidth = child.measuredWidth + lp.leftMargin + lp.rightMargin
            childMaxHeight =
                max(childMaxHeight, child.measuredHeight + lp.topMargin + lp.bottomMargin)
            if (childWidth > mMaxWidth) {
                if (curLineWidth > 0) {
                    linePosition++
                }
                curLineWidth = 0
            } else {
                if (curLineWidth + childWidth > mMaxWidth) {
                    linePosition++
                    curLineWidth = childWidth
                } else {
                    curLineWidth += childWidth
                }
            }
            if (linePosition == 0) {
                firstLineChildWidth += childWidth
                firstLineChildCount++
            }
            child.setTag(R.id.wrap_radio_button_line, linePosition)
        }

        if (linePosition == 0) {
            if (firstLineChildCount > 1) {
                firstLinePadding = (mMaxWidth - firstLineChildWidth) / (firstLineChildCount - 1)
            }
        } else {
            firstLinePadding = 0
        }

        val measuredWidth = widthMeasureSpec
        val measuredHeight = MeasureSpec.makeMeasureSpec(
            (linePosition + 1) * childMaxHeight + paddingTop + paddingBottom,
            heightMode
        )

        super.onMeasure(widthMeasureSpec, measuredHeight)
    }

    override fun onLayout(changed: Boolean, l: Int, t: Int, r: Int, b: Int) {
        val childCount = getChildCount()

        val x = getPaddingLeft()
        val y = paddingTop
        var left = x
        var currentLine = 0
        for (i in 0..<childCount) {
            val child = getChildAt(i)
            if (child.visibility != View.VISIBLE) {
                continue
            }
            val lp = child.layoutParams as LayoutParams
            val childWidth = child.measuredWidth + lp.leftMargin + lp.rightMargin

            val linePosition = child.getTag(R.id.wrap_radio_button_line) as Int
            if (currentLine != linePosition) {
                currentLine = linePosition
                left = x
            }
            val top = y + linePosition * childMaxHeight
            child.layout(left, top, left + childWidth, top + childMaxHeight)
            left += childWidth + firstLinePadding
        }
    }

    companion object {
        private const val TAG = "WrapRadioGroup"
    }
}
