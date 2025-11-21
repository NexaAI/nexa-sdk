package com.nexa.demo.activity

import android.app.Activity
import android.content.Intent
import android.graphics.Color
import android.os.Bundle
import android.util.Log
import android.view.View
import com.gyf.immersionbar.ktx.immersionBar
import com.nexa.demo.adapter.ShowFileDirAdapter
import com.nexa.demo.bean.IndexSelectedDirBean
import com.nexa.demo.databinding.ActivityFolderBinding
import com.nexa.demo.utils.inflate
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.launch
import java.io.File

class FolderActivity : Activity() {

    private val binding by inflate<ActivityFolderBinding>()
    private val rootDir = File("/sdcard")
    private lateinit var adapter: ShowFileDirAdapter
    private lateinit var job: Job

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        immersionBar {
            statusBarColorInt(Color.WHITE)
            statusBarDarkFont(true)
            fitsSystemWindows(true)
        }
        binding.btnBack.setOnClickListener {
            finish()
        }
        binding.btnImport.setOnClickListener {
            adapter.getSelectedImages().let { images ->
                setResult(RESULT_OK, Intent().apply {
                    Log.d("nfl", "return all images:$images")
                    this.putStringArrayListExtra(KEY_SELECT_IMAGES, images)
                })
            }
            finish()
        }
        job = CoroutineScope(Dispatchers.IO).launch {
            val showDirs =
                rootDir.listFiles()!!
                    .filter { file ->
                        file.isDirectory && !file.name.startsWith(".") &&
                                file.name != "Android" && file.name != "Alarms"
                    }
                    .sorted()
                    .map {
                        IndexSelectedDirBean(it)
                    }
            showDirs.forEach {
                initShowDirsData(it.dir, it)
            }
            runOnUiThread {
                binding.flLoading.visibility = View.GONE
                adapter = ShowFileDirAdapter(showDirs)
                binding.rvFiles.adapter = adapter
            }
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        job.cancel()
    }

    private fun initShowDirsData(dir: File, bean: IndexSelectedDirBean) {
        if (bean.subImageFiles == null) {
            bean.subImageFiles = arrayListOf()
        }
        if (bean.subVideoFiles == null) {
            bean.subVideoFiles = arrayListOf()
        }
        dir.listFiles()?.forEach { subFile ->
            if (subFile.isFile) {
                if (subFile.name.endsWith("jpg", true) ||
                    subFile.name.endsWith("jpeg", true) ||
                    subFile.name.endsWith("png", true)
                ) {
                    bean.subImageFiles!!.add(subFile.absolutePath)
                } else if (subFile.name.endsWith("mp4", true)) {
                    bean.subVideoFiles!!.add(subFile.absolutePath)
                }
            } else {
                initShowDirsData(subFile, bean)
            }
        }
    }

    companion object {
        const val KEY_SELECT_DIRS = "select_dirs"
        const val KEY_SELECT_IMAGES = "select_dirs"
    }
}