use std::env::current_dir;
use std::fs;
use std::fs::File;
use std::path::PathBuf;
use serde::{Deserialize, Serialize};
use std::time::SystemTime;
use crate::constants;
use crate::errors::uninitialized_error;

#[derive(Serialize, Deserialize)]
pub struct ProjectMetadata {
    pub creation_timestamp: SystemTime
}

#[derive(Serialize, Deserialize, Clone)]
pub struct TrackedFile {
    file_path: PathBuf,
    hash: String
}

impl TrackedFile {
    fn track(file: &PathBuf) -> anyhow::Result<TrackedFile> {
        let hash = calculate_hash(&file)?;

        Ok(TrackedFile {
            file_path: file.clone(),
            hash,
        })
    }


}

fn calculate_hash(file: &PathBuf) -> anyhow::Result<String> {
    let bytes = fs::read(&file)?;
    let hash = sha256::digest(&bytes);
    Ok(hash)
}
#[derive(Serialize, Deserialize, Clone)]
pub struct ProjectTrackedFilesData {
    pub tracked_files: Vec<TrackedFile>
}

pub struct FileTracking {
    pub src: PathBuf
}

impl FileTracking {
    pub fn init() -> anyhow::Result<FileTracking> {
        let aether_dir = current_dir()?.join(constants::AETHER_DIR);

        if !aether_dir.exists() {
            return Err(uninitialized_error()?)
        }

        let file = aether_dir.join(constants::TRACKED_FILES_FILE);

        if !file.exists() {

            let data = ProjectTrackedFilesData {
                tracked_files: vec![]
            };

            fs::write(&file, serde_json::to_vec(&data)?)?;
        }


        Ok(FileTracking {
            src: file
        })
    }

    pub fn track(&self, file: &PathBuf) -> anyhow::Result<()> {
        let mut tracking_data: ProjectTrackedFilesData = serde_json::from_reader(File::open(&self.src)?)?;

        let cloned = tracking_data.clone();
        let existing: Vec<&TrackedFile> = cloned.tracked_files.iter().filter(|f| f.file_path.to_str() == file.to_str()).collect();

        if !existing.is_empty() {
            let existing_file = existing.first().unwrap();


            if existing_file.hash == calculate_hash(file)? {
                println!("File {} is unchanged", existing_file.file_path.display());
                return Ok(())
            }

            tracking_data.tracked_files.retain(|f| f.file_path != existing_file.file_path);
        }


        tracking_data.tracked_files.push(TrackedFile::track(file)?);

        fs::write(&self.src, serde_json::to_vec(&tracking_data)?)?;

        Ok(())
    }


}
